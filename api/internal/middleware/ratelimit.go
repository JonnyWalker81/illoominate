package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/fulldisclosure/api/internal/auth"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string]*rateLimitEntry
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
	cleanup  time.Duration
}

type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Limit   int           // Max requests per window
	Window  time.Duration // Time window
	Cleanup time.Duration // Cleanup interval for expired entries
}

// DefaultRateLimitConfig returns default rate limit settings
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:   300,
		Window:  time.Minute,
		Cleanup: 5 * time.Minute,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateLimitEntry),
		limit:    config.Limit,
		window:   config.Window,
		cleanup:  config.Cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop periodically removes expired entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.After(entry.windowEnd) {
				delete(rl.requests, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// Allow checks if a request should be allowed and increments the counter
func (rl *RateLimiter) Allow(key string) (bool, int, time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	entry, exists := rl.requests[key]

	if !exists || now.After(entry.windowEnd) {
		// New window
		rl.requests[key] = &rateLimitEntry{
			count:     1,
			windowEnd: now.Add(rl.window),
		}
		return true, rl.limit - 1, now.Add(rl.window)
	}

	if entry.count >= rl.limit {
		return false, 0, entry.windowEnd
	}

	entry.count++
	return true, rl.limit - entry.count, entry.windowEnd
}

// RateLimit creates rate limiting middleware
func RateLimit(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			allowed, remaining, resetTime := limiter.Allow(key)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				log.Warn().
					Str("key", key).
					Int("limit", limiter.limit).
					Msg("Rate limit exceeded")

				w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFunc returns the client IP as the rate limit key
func IPKeyFunc(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// UserKeyFunc returns the user ID as the rate limit key
func UserKeyFunc(r *http.Request) string {
	userID, ok := auth.UserIDFromContext(r.Context())
	if ok {
		return "user:" + userID.String()
	}
	// Fall back to IP
	return "ip:" + IPKeyFunc(r)
}

// SDKProjectKeyFunc returns the SDK project ID as the rate limit key
func SDKProjectKeyFunc(r *http.Request) string {
	projectID, ok := auth.SDKProjectFromContext(r.Context())
	if ok {
		return "sdk:" + projectID.String()
	}
	// Fall back to IP
	return "ip:" + IPKeyFunc(r)
}

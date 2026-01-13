package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidSDKToken  = errors.New("invalid SDK token")
	ErrSDKTokenExpired  = errors.New("SDK token expired")
	ErrOriginNotAllowed = errors.New("origin not allowed for this token")
	ErrMissingSDKToken  = errors.New("missing SDK token")
)

// SDKTokenValidator validates SDK tokens for anonymous SDK submissions
type SDKTokenValidator struct {
	db *pgxpool.Pool
}

// SDKToken represents an SDK token from the database
type SDKToken struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	TokenHash      string
	AllowedOrigins []string
	RateLimit      int
	LastUsedAt     *time.Time
	ExpiresAt      *time.Time
	CreatedAt      time.Time
}

// NewSDKTokenValidator creates a new SDK token validator
func NewSDKTokenValidator(db *pgxpool.Pool) *SDKTokenValidator {
	return &SDKTokenValidator{db: db}
}

// ValidateToken validates an SDK token and returns the associated project ID
func (v *SDKTokenValidator) ValidateToken(ctx context.Context, token string, origin string) (uuid.UUID, error) {
	// Hash the token for lookup
	tokenHash := hashToken(token)

	// Look up the token in the database
	query := `
		SELECT id, project_id, name, allowed_origins, rate_limit, last_used_at, expires_at, created_at
		FROM sdk_tokens
		WHERE token_hash = $1
	`

	var sdkToken SDKToken
	var allowedOriginsJSON []byte
	var lastUsedAt, expiresAt sql.NullTime

	err := v.db.QueryRow(ctx, query, tokenHash).Scan(
		&sdkToken.ID,
		&sdkToken.ProjectID,
		&sdkToken.Name,
		&allowedOriginsJSON,
		&sdkToken.RateLimit,
		&lastUsedAt,
		&expiresAt,
		&sdkToken.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrInvalidSDKToken
		}
		return uuid.Nil, fmt.Errorf("failed to lookup SDK token: %w", err)
	}

	// Check expiration
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return uuid.Nil, ErrSDKTokenExpired
	}

	// Check origin if allowed origins are configured
	if len(allowedOriginsJSON) > 0 && origin != "" {
		if !isOriginAllowed(allowedOriginsJSON, origin) {
			return uuid.Nil, ErrOriginNotAllowed
		}
	}

	// Update last used timestamp (non-blocking)
	go v.updateLastUsed(context.Background(), sdkToken.ID)

	return sdkToken.ProjectID, nil
}

// updateLastUsed updates the last_used_at timestamp for a token
func (v *SDKTokenValidator) updateLastUsed(ctx context.Context, tokenID uuid.UUID) {
	query := `UPDATE sdk_tokens SET last_used_at = NOW() WHERE id = $1`
	_, _ = v.db.Exec(ctx, query, tokenID)
}

// GenerateToken generates a new SDK token
func GenerateToken() (string, string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Encode as hex with prefix
	token := "sdk_" + hex.EncodeToString(bytes)
	hash := hashToken(token)

	return token, hash, nil
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(allowedOriginsJSON []byte, origin string) bool {
	// Parse JSON array
	var allowedOrigins []string
	// Simple JSON parsing - in production use encoding/json
	str := string(allowedOriginsJSON)
	str = strings.TrimPrefix(str, "[")
	str = strings.TrimSuffix(str, "]")
	if str == "" {
		return true // Empty list means all origins allowed
	}

	parts := strings.Split(str, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		allowedOrigins = append(allowedOrigins, part)
	}

	// Check if origin matches any allowed origin
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if matchOrigin(allowed, origin) {
			return true
		}
	}

	return false
}

// matchOrigin checks if an origin matches a pattern
// Supports wildcards like "*.example.com"
func matchOrigin(pattern, origin string) bool {
	if pattern == origin {
		return true
	}

	// Handle wildcard subdomains
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // Remove the "*"
		return strings.HasSuffix(origin, suffix)
	}

	return false
}

// ExtractSDKTokenFromHeader extracts the SDK token from the request header
func ExtractSDKTokenFromHeader(r *http.Request) (string, error) {
	token := r.Header.Get("X-SDK-Token")
	if token == "" {
		return "", ErrMissingSDKToken
	}
	return token, nil
}

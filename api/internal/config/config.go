package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    int    `env:"PORT,default=8080"`
	Env     string `env:"ENV,default=development"`
	LogLevel string `env:"LOG_LEVEL,default=info"`

	// CORS
	AllowedOrigins []string `env:"ALLOWED_ORIGINS,default=http://localhost:5173"`

	// Supabase
	SupabaseURL       string `env:"SUPABASE_URL,required"`
	SupabaseAnonKey   string `env:"SUPABASE_ANON_KEY,required"`
	SupabaseServiceKey string `env:"SUPABASE_SERVICE_KEY,required"`
	SupabaseJWTSecret string `env:"SUPABASE_JWT_SECRET,required"`

	// Database
	DatabaseURL        string        `env:"DATABASE_URL,required"`
	DBMaxOpenConns     int           `env:"DB_MAX_OPEN_CONNS,default=25"`
	DBMaxIdleConns     int           `env:"DB_MAX_IDLE_CONNS,default=5"`
	DBConnMaxLifetime  time.Duration `env:"DB_CONN_MAX_LIFETIME,default=5m"`

	// Google Cloud Storage
	GCSBucket           string        `env:"GCS_BUCKET,default=fulldisclosure-attachments-dev"`
	GCSProjectID        string        `env:"GCS_PROJECT_ID"`
	GCSUploadURLExpiry  time.Duration `env:"GCS_UPLOAD_URL_EXPIRY,default=15m"`
	GCSDownloadURLExpiry time.Duration `env:"GCS_DOWNLOAD_URL_EXPIRY,default=1h"`

	// Rate Limiting
	RateLimitEnabled     bool   `env:"RATE_LIMIT_ENABLED,default=true"`
	RedisURL             string `env:"REDIS_URL"`
	RateLimitSDKIdentify int    `env:"RATE_LIMIT_SDK_IDENTIFY,default=100"`
	RateLimitSDKFeedback int    `env:"RATE_LIMIT_SDK_FEEDBACK,default=10"`
	RateLimitVote        int    `env:"RATE_LIMIT_VOTE,default=30"`
	RateLimitComment     int    `env:"RATE_LIMIT_COMMENT,default=10"`
	RateLimitGeneral     int    `env:"RATE_LIMIT_GENERAL,default=300"`

	// SDK Token
	SDKTokenSecret string        `env:"SDK_TOKEN_SECRET,required"`
	SDKTokenExpiry time.Duration `env:"SDK_TOKEN_EXPIRY,default=24h"`
}

// Load loads configuration from environment variables
func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}

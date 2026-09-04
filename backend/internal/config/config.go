// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the backend.
type Config struct {
	DatabaseURL         string
	JWTAccessSecret     string
	JWTRefreshSecret    string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	FileStorage         string // "local" or "s3"
	LocalStoragePath    string
	S3Bucket            string
	S3Region            string
	S3AccessKey         string
	S3SecretKey         string
	S3Endpoint          string
	AppEnv              string
	HTTPAddr            string
	CORSAllowedOrigins  []string
	RateLimitRPS        int
}

// Load reads configuration from the process environment.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://osint:osint@localhost:5432/osint?sslmode=disable"),
		JWTAccessSecret:    getEnv("JWT_ACCESS_SECRET", "dev-access-secret-change-me"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me"),
		FileStorage:        getEnv("FILE_STORAGE", "local"),
		LocalStoragePath:   getEnv("LOCAL_STORAGE_PATH", "./uploads"),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		S3Region:           getEnv("S3_REGION", ""),
		S3AccessKey:        getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:        getEnv("S3_SECRET_KEY", ""),
		S3Endpoint:         getEnv("S3_ENDPOINT", ""),
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")),
		RateLimitRPS:       getEnvInt("RATE_LIMIT_RPS", 20),
	}

	var err error
	cfg.AccessTokenTTL, err = time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}
	cfg.RefreshTokenTTL, err = time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}

	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT secrets must be set")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	out := []string{}
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

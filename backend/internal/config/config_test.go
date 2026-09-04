package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear relevant env vars to test defaults.
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("expected default access TTL 15m, got %v", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 168*time.Hour {
		t.Errorf("expected default refresh TTL 168h, got %v", cfg.RefreshTokenTTL)
	}
	if cfg.RateLimitRPS != 20 {
		t.Errorf("expected default rate limit 20, got %d", cfg.RateLimitRPS)
	}
}

func TestLoadCustomValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_RPS", "50")
	os.Setenv("ACCESS_TOKEN_TTL", "30m")
	defer func() {
		os.Unsetenv("RATE_LIMIT_RPS")
		os.Unsetenv("ACCESS_TOKEN_TTL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.RateLimitRPS != 50 {
		t.Errorf("expected rate limit 50, got %d", cfg.RateLimitRPS)
	}
	if cfg.AccessTokenTTL != 30*time.Minute {
		t.Errorf("expected access TTL 30m, got %v", cfg.AccessTokenTTL)
	}
}

package config

import (
	"os"
	"testing"
)

func TestLoad_RequiredEnvVars(t *testing.T) {
	// Set required vars
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	cfg := Load()

	if cfg.DatabaseURL != "postgresql://test:test@localhost/test" {
		t.Errorf("expected DATABASE_URL to be set, got %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWT_SECRET to be set, got %q", cfg.JWTSecret)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("expected default Port 3000, got %q", cfg.Port)
	}
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("expected default RedisURL, got %q", cfg.RedisURL)
	}
	if cfg.S3Region != "ru-central1" {
		t.Errorf("expected default S3Region ru-central1, got %q", cfg.S3Region)
	}
	if cfg.AppBaseURL != "https://repa.app" {
		t.Errorf("expected default AppBaseURL, got %q", cfg.AppBaseURL)
	}
	if cfg.DevMode {
		t.Error("expected DevMode to be false by default")
	}
}

func TestLoad_DevMode(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("DEV_MODE", "true")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DEV_MODE")

	cfg := Load()

	if !cfg.DevMode {
		t.Error("expected DevMode to be true when DEV_MODE=true")
	}
}

func TestLoad_DevModeNotSetInProd(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Unsetenv("DEV_MODE")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	cfg := Load()

	if cfg.DevMode {
		t.Error("DevMode must be false when DEV_MODE is not set — OTP codes would leak in production")
	}
}

func TestMustEnv_Panics(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR_FOR_TEST")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected mustEnv to panic on missing required var")
		}
	}()

	mustEnv("NONEXISTENT_VAR_FOR_TEST")
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR_FOR_TEST")

	val := getEnv("NONEXISTENT_VAR_FOR_TEST", "fallback-value")
	if val != "fallback-value" {
		t.Errorf("expected fallback, got %q", val)
	}
}

func TestGetEnv_Override(t *testing.T) {
	os.Setenv("TEST_OVERRIDE_VAR", "custom")
	defer os.Unsetenv("TEST_OVERRIDE_VAR")

	val := getEnv("TEST_OVERRIDE_VAR", "fallback")
	if val != "custom" {
		t.Errorf("expected custom, got %q", val)
	}
}

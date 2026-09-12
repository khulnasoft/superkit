package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("expected Env=development, got %s", cfg.Env)
	}
	if cfg.Listen != ":3000" {
		t.Errorf("expected Listen=:3000, got %s", cfg.Listen)
	}
	if cfg.DBDriver != "sqlite3" {
		t.Errorf("expected DBDriver=sqlite3, got %s", cfg.DBDriver)
	}
	if cfg.DBName != "app_db" {
		t.Errorf("expected DBName=app_db, got %s", cfg.DBName)
	}
	if cfg.AuthRedirectAfterLogin != "/profile" {
		t.Errorf("expected redirect=/profile, got %s", cfg.AuthRedirectAfterLogin)
	}
	if cfg.AuthSessionExpiryHours != 48 {
		t.Errorf("expected session expiry=48, got %d", cfg.AuthSessionExpiryHours)
	}
	if cfg.AuthVerificationExpiryHours != 1 {
		t.Errorf("expected verification expiry=1, got %d", cfg.AuthVerificationExpiryHours)
	}
}

func TestLoadOverrides(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	os.Setenv("SUPERKIT_ENV", "production")
	os.Setenv("HTTP_LISTEN_ADDR", ":8080")
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("SUPERKIT_AUTH_REDIRECT_AFTER_LOGIN", "/dashboard")
	os.Setenv("SUPERKIT_AUTH_SESSION_EXPIRY_IN_HOURS", "24")
	os.Setenv("SUPERKIT_AUTH_EMAIL_VERIFICATION_EXPIRY_IN_HOURS", "2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("expected Env=production, got %s", cfg.Env)
	}
	if cfg.Listen != ":8080" {
		t.Errorf("expected Listen=:8080, got %s", cfg.Listen)
	}
	if cfg.DBName != "test_db" {
		t.Errorf("expected DBName=test_db, got %s", cfg.DBName)
	}
	if cfg.AuthRedirectAfterLogin != "/dashboard" {
		t.Errorf("expected redirect=/dashboard, got %s", cfg.AuthRedirectAfterLogin)
	}
	if cfg.AuthSessionExpiryHours != 24 {
		t.Errorf("expected session expiry=24, got %d", cfg.AuthSessionExpiryHours)
	}
	if cfg.AuthVerificationExpiryHours != 2 {
		t.Errorf("expected verification expiry=2, got %d", cfg.AuthVerificationExpiryHours)
	}
}

func TestValidateMissingSecret(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestValidateShortSecret(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "short")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short secret")
	}
}

func TestValidateUnsupportedDriver(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	os.Setenv("DB_DRIVER", "mysql")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestGetenv(t *testing.T) {
	os.Setenv("FOO", "bar")
	if got := getenv("FOO", "default"); got != "bar" {
		t.Errorf("expected bar, got %s", got)
	}
	if got := getenv("MISSING", "default"); got != "default" {
		t.Errorf("expected default, got %s", got)
	}
}

func TestLoadEmptyDatabaseNameDefaults(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", "")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "app_db", cfg.DBName)
}

func TestLoadInvalidEnv(t *testing.T) {
	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	os.Setenv("SUPERKIT_ENV", "invalid")

	_, err := Load()
	assert.Error(t, err)
}

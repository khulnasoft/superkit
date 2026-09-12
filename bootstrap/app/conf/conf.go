package conf

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env    string
	Listen string

	DBDriver   string
	DBName     string
	DBUser     string
	DBPassword string
	DBHost     string

	MigrationDir string

	Secret                     string
	AuthRedirectAfterLogin     string
	AuthSessionExpiryHours     int
	AuthSkipVerify             bool
	AuthVerificationExpiryHours int
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:    getenv("SUPERKIT_ENV", "development"),
		Listen: getenv("HTTP_LISTEN_ADDR", ":3000"),

		DBDriver:   getenv("DB_DRIVER", "sqlite3"),
		DBName:     getenv("DB_NAME", "app_db"),
		DBUser:     getenv("DB_USER", ""),
		DBPassword: getenv("DB_PASSWORD", ""),
		DBHost:     getenv("DB_HOST", ""),

		MigrationDir: getenv("MIGRATION_DIR", "app/db/migrations"),

		Secret:                     getenv("SUPERKIT_SECRET", ""),
		AuthRedirectAfterLogin:     getenv("SUPERKIT_AUTH_REDIRECT_AFTER_LOGIN", "/profile"),
		AuthSkipVerify:             getenv("SUPERKIT_AUTH_SKIP_VERIFY", "false") == "true",
	}

	expiry, err := strconv.Atoi(getenv("SUPERKIT_AUTH_SESSION_EXPIRY_IN_HOURS", "48"))
	if err != nil {
		expiry = 48
	}
	cfg.AuthSessionExpiryHours = expiry

	verificationExpiry, err := strconv.Atoi(getenv("SUPERKIT_AUTH_EMAIL_VERIFICATION_EXPIRY_IN_HOURS", "1"))
	if err != nil {
		verificationExpiry = 1
	}
	cfg.AuthVerificationExpiryHours = verificationExpiry

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	switch c.Env {
	case "development", "production", "test":
	default:
		return fmt.Errorf("invalid SUPERKIT_ENV: %s (must be development, production, or test)", c.Env)
	}

	if len(c.Secret) < 32 {
		return fmt.Errorf("SUPERKIT_SECRET must be at least 32 characters, got %d", len(c.Secret))
	}

	switch c.DBDriver {
	case "sqlite3":
		if c.DBName == "" {
			c.DBName = "app_db"
		}
	default:
		return fmt.Errorf("unsupported database driver: %s (only sqlite3 is supported)", c.DBDriver)
	}

	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

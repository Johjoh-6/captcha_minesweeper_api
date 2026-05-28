package servers

import (
	"captcha_sweeper/internal/env"
	"fmt"
)

type Config struct {
	BaseURL   string
	HttpPort  int
	BasicAuth struct {
		Username       string
		HashedPassword string
	}
	Cookie struct {
		Name      string
		SecretKey string
	}
	DB struct {
		Dsn string
	}
}

func LoadConfig() *Config {
	cfg := Config{}

	// Load configuration from environment variables
	cfg.BaseURL = env.GetString("BASE_URL", "http://localhost:8080")
	cfg.HttpPort = env.GetInt("HTTP_PORT", 8080)
	cfg.BasicAuth.Username = env.GetString("BASIC_AUTH_USER", "")
	cfg.BasicAuth.HashedPassword = env.GetString("BASIC_AUTH_PASSWORD", "")
	cfg.Cookie.Name = env.GetString("COOKIE_NAME", "captcha_session")
	cfg.Cookie.SecretKey = env.GetString("COOKIE_SECRET_KEY", "")
	cfg.DB.Dsn = env.GetString("DB_DSN", "")

	return &cfg
}

// Validate checks that all required configuration fields are present and valid
func (c *Config) Validate() error {
	if c.DB.Dsn == "" {
		return fmt.Errorf("config validation failed: DB_DSN is required")
	}

	if c.Cookie.SecretKey == "" {
		return fmt.Errorf("config validation failed: COOKIE_SECRET_KEY is required")
	}

	if c.BasicAuth.Username == "" {
		return fmt.Errorf("config validation failed: BASIC_AUTH_USER is required")
	}

	if c.BasicAuth.HashedPassword == "" {
		return fmt.Errorf("config validation failed: BASIC_AUTH_PASSWORD is required")
	}

	return nil
}

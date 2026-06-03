package servers

import (
	"captcha_sweeper/internal/env"
	"fmt"
	"time"
)

type IdentifierMode string

const (
	IdentifierModeCookie  IdentifierMode = "cookie"
	IdentifierModeSession IdentifierMode = "session"
	IdentifierModeJWT     IdentifierMode = "jwt"
)

type Config struct {
	BaseURL   string
	HttpPort  int
	BasicAuth struct {
		Username       string
		HashedPassword string
	}
	DB struct {
		Dsn string
	}
	Identifier IdentifierMode
	Cookie     struct {
		Name      string
		SecretKey string
	}
	JWT struct {
		SecretKey      string
		ExpiryDuration time.Duration
	}
	SessionHeaderName string
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
	cfg.Identifier = IdentifierMode(env.GetString("IDENTIFIER", "session"))
	cfg.JWT.SecretKey = env.GetString("JWT_SECRET_KEY", "")
	cfg.JWT.ExpiryDuration = time.Hour * time.Duration(env.GetInt("JWT_EXPIRY_DURATION", 1))
	cfg.SessionHeaderName = env.GetString("SESSION_HEADER_NAME", "X-Session-ID")

	return &cfg
}

// Validate checks that all required configuration fields are present and valid
func (c *Config) Validate() error {
	if c.DB.Dsn == "" {
		return fmt.Errorf("config validation failed: DB_DSN is required")
	}

	if c.BasicAuth.Username == "" {
		return fmt.Errorf("config validation failed: BASIC_AUTH_USER is required")
	}

	if c.BasicAuth.HashedPassword == "" {
		return fmt.Errorf("config validation failed: BASIC_AUTH_PASSWORD is required")
	}

	switch c.Identifier {
	case IdentifierModeCookie, IdentifierModeSession, IdentifierModeJWT:
		// Valid, do nothing
	case "":
		return fmt.Errorf("config validation failed: IDENTIFIER is required")
	default:
		return fmt.Errorf("config validation failed: IDENTIFIER must be 'cookie', 'session', or 'jwt', got '%s'", c.Identifier)
	}

	return nil
}

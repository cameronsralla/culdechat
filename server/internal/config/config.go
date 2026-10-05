// Package config loads and validates process configuration from the environment.
// Deploy-time values live here; operator-editable runtime values live in the
// settings table (see internal/settings).
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type Config struct {
	Env        Env
	ListenAddr string
	PublicURL  string // https://square.example.org — used in emails and CORS default

	DatabaseURL string
	DBMaxConns  int32

	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	CORSOrigins    []string
	TrustedProxies []string // CIDRs/IPs allowed to set X-Forwarded-For (Caddy)

	RequestTimeout time.Duration
	MaxJSONBody    int64
	MaxUploadBody  int64

	RateLimitEnabled bool

	SMTP SMTP

	BootstrapAdmin BootstrapAdmin

	MediaDir string
}

type SMTP struct {
	Host     string
	Port     int
	From     string
	User     string
	Password string
	StartTLS bool
}

type BootstrapAdmin struct {
	Email    string
	Password string
	Name     string
	Unit     string
}

func (c Config) IsProd() bool { return c.Env == EnvProd }

// Load reads configuration from the environment and validates it.
func Load() (Config, error) {
	c := Config{
		Env:              Env(get("APP_ENV", "dev")),
		ListenAddr:       get("LISTEN_ADDR", ":8080"),
		PublicURL:        strings.TrimRight(get("PUBLIC_URL", "http://localhost:5173"), "/"),
		DatabaseURL:      get("DATABASE_URL", ""),
		DBMaxConns:       int32(getInt("DB_MAX_CONNS", 10)),
		JWTSecret:        []byte(get("JWT_SECRET", "")),
		AccessTokenTTL:   getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:  getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		CORSOrigins:      getList("CORS_ORIGINS"),
		TrustedProxies:   getList("TRUSTED_PROXIES"),
		RequestTimeout:   getDuration("REQUEST_TIMEOUT", 30*time.Second),
		MaxJSONBody:      int64(getInt("MAX_JSON_BODY_BYTES", 1<<20)),
		MaxUploadBody:    int64(getInt("MAX_UPLOAD_BODY_BYTES", 20<<20)),
		RateLimitEnabled: getBool("RATE_LIMIT", true),
		MediaDir:         get("MEDIA_DIR", "./data/media"),
		SMTP: SMTP{
			Host:     get("SMTP_HOST", ""),
			Port:     getInt("SMTP_PORT", 587),
			From:     get("SMTP_FROM", ""),
			User:     get("SMTP_USER", ""),
			Password: get("SMTP_PASS", ""),
			StartTLS: getBool("SMTP_STARTTLS", true),
		},
		BootstrapAdmin: BootstrapAdmin{
			Email:    strings.ToLower(strings.TrimSpace(get("BOOTSTRAP_ADMIN_EMAIL", ""))),
			Password: get("BOOTSTRAP_ADMIN_PASSWORD", ""),
			Name:     get("BOOTSTRAP_ADMIN_NAME", "Admin"),
			Unit:     get("BOOTSTRAP_ADMIN_UNIT", "Admin"),
		},
	}

	if len(c.CORSOrigins) == 0 {
		c.CORSOrigins = []string{c.PublicURL}
	}

	return c, c.validate()
}

func (c Config) validate() error {
	var errs []error
	if c.Env != EnvDev && c.Env != EnvProd {
		errs = append(errs, fmt.Errorf("APP_ENV must be dev or prod, got %q", c.Env))
	}
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 bytes"))
	}
	if c.IsProd() {
		if !strings.HasPrefix(c.PublicURL, "https://") {
			errs = append(errs, errors.New("PUBLIC_URL must be https in prod"))
		}
		for _, o := range c.CORSOrigins {
			if o == "*" {
				errs = append(errs, errors.New("CORS_ORIGINS may not contain * in prod"))
			}
		}
		if c.SMTP.Host == "" || c.SMTP.From == "" {
			errs = append(errs, errors.New("SMTP_HOST and SMTP_FROM are required in prod (invites)"))
		}
	}
	return errors.Join(errs...)
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func getInt(key string, def int) int {
	if v := get(key, ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	switch strings.ToLower(get(key, "")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := get(key, ""); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getList(key string) []string {
	raw := get(key, "")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

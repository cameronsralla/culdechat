package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cameronsralla/culdechat/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

type Config struct {
	DatabaseURL string
	Host        string
	Port        int
	Database    string
	User        string
	Password    string
	SSLMode     string
	MinConns    int32
	MaxConns    int32
	Timeout     time.Duration
}

func readEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func readIntEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func readInt32Env(key string, def int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return int32(n)
		}
	}
	return def
}

func readDurationSecondsEnv(key string, defSeconds int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(defSeconds) * time.Second
}

func loadConfig() Config {
	return Config{
		DatabaseURL: readEnv("DATABASE_URL", ""),
		Host:        readEnv("PGHOST", "localhost"),
		Port:        readIntEnv("PGPORT", 5432),
		Database:    readEnv("PGDATABASE", "culdechat"),
		User:        readEnv("PGUSER", "postgres"),
		Password:    readEnv("PGPASSWORD", ""),
		SSLMode:     readEnv("PGSSLMODE", "disable"),
		MinConns:    readInt32Env("PGPOOL_MIN_CONNS", 0),
		MaxConns:    readInt32Env("PGPOOL_MAX_CONNS", 10),
		Timeout:     readDurationSecondsEnv("PGCONNECT_TIMEOUT", 5),
	}
}

func dsnFromConfig(c Config) string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	// Construct DSN. Avoid logging sensitive info elsewhere.
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.User), url.QueryEscape(c.Password), c.Host, c.Port, c.Database, c.SSLMode,
	)
}

// Initialize opens the pgx connection pool using env configuration.
func Initialize(ctx context.Context) (*pgxpool.Pool, error) {
	if pool != nil {
		return pool, nil
	}

	cfg := loadConfig()
	dsn := dsnFromConfig(cfg)

	parsed, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		utils.Errorf("pgx pool parse config failed: %v", err)
		return nil, err
	}

	parsed.MinConns = cfg.MinConns
	parsed.MaxConns = cfg.MaxConns
	// Connect timeout in v5 is on ConnConfig
	parsed.ConnConfig.ConnectTimeout = cfg.Timeout

	p, err := pgxpool.NewWithConfig(ctx, parsed)
	if err != nil {
		utils.Errorf("pgx pool connect failed: %v", err)
		return nil, err
	}

	pool = p
	utils.Infof("connected to Postgres at %s:%d db=%s (min=%d max=%d)", cfg.Host, cfg.Port, cfg.Database, cfg.MinConns, cfg.MaxConns)
	return pool, nil
}

// Pool returns the initialized pool or nil if Initialize hasn't been called.
func Pool() *pgxpool.Pool { return pool }

// Close closes the pool if initialized.
func Close() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
}

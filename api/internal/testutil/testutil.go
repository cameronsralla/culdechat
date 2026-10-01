package testutil

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/models"
	"github.com/jackc/pgx/v5"
)

// Setup initializes a dedicated test database, migrates schema, and truncates tables.
func Setup(t *testing.T) {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("CULDECHAT_RATE_LIMIT", "off")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASS")
	os.Setenv("JWT_ISSUER", "culdechat-test")
	os.Setenv("PGSSLMODE", getenv("PGSSLMODE", "disable"))
	os.Setenv("PGHOST", getenv("PGHOST", "localhost"))
	os.Setenv("PGPORT", getenv("PGPORT", "5432"))
	os.Setenv("PGUSER", getenv("PGUSER", "postgres"))
	os.Setenv("PGPASSWORD", getenv("PGPASSWORD", "postgres"))
	os.Setenv("PGDATABASE", getenv("TEST_PGDATABASE", "culdechat_test"))

	ctx := context.Background()
	if err := ensureDatabase(ctx); err != nil {
		t.Fatalf("postgres is required for API tests (start infra/dev/docker-compose.yml): %v", err)
	}

	if _, err := postgres.Initialize(ctx); err != nil {
		t.Fatalf("postgres init: %v", err)
	}
	if err := models.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	Reset(t)
}

// Reset truncates all application tables.
func Reset(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := postgres.Pool().Exec(ctx, `
TRUNCATE TABLE messages, conversations, refresh_tokens, post_reactions, comments, posts, board_subscriptions, boards, users CASCADE;
`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func ensureDatabase(ctx context.Context) error {
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	pass := os.Getenv("PGPASSWORD")
	ssl := os.Getenv("PGSSLMODE")
	db := os.Getenv("PGDATABASE")

	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		url.QueryEscape(user), url.QueryEscape(pass), host, port, ssl)
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, _ = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", db))
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

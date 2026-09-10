package models

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
)

// Migrate creates or updates all application tables.
func Migrate(ctx context.Context) error {
	if err := EnsureUsersTable(ctx); err != nil {
		return err
	}
	if err := EnsureBoardsTable(ctx); err != nil {
		return err
	}
	if err := EnsureBoardSubscriptionsTable(ctx); err != nil {
		return err
	}
	if err := EnsurePostsTable(ctx); err != nil {
		return err
	}
	if err := EnsureCommentsTable(ctx); err != nil {
		return err
	}
	if err := EnsurePostReactionsTable(ctx); err != nil {
		return err
	}
	if err := EnsureRefreshTokensTable(ctx); err != nil {
		return err
	}
	return nil
}

// BootstrapAdmin creates an active admin from env vars when none exists.
// Used so local/dev can exercise admin-only onboarding without a mailer.
func BootstrapAdmin(ctx context.Context) error {
	email := strings.TrimSpace(strings.ToLower(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	existing, err := GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	admins, err := CountAdmins(ctx)
	if err != nil {
		return err
	}
	if admins > 0 {
		return nil
	}

	unit := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_UNIT"))
	if unit == "" {
		unit = "Admin"
	}
	name := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_NAME"))
	if name == "" {
		name = "Business Admin"
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := &User{
		UnitNumber:     unit,
		Email:          email,
		Name:           name,
		HashedPassword: hashed,
		IsAdmin:        true,
		Status:         UserStatusActive,
	}
	if err := InsertUser(ctx, user); err != nil {
		return err
	}
	utils.Infof("bootstrapped admin email=%s id=%s", user.Email, user.ID)
	return nil
}

func tableExists(ctx context.Context, name string) (bool, error) {
	pool := postgres.Pool()
	if pool == nil {
		return false, errors.New("postgres pool is not initialized")
	}
	var exists bool
	err := pool.QueryRow(ctx, `
SELECT EXISTS (
	SELECT 1 FROM information_schema.tables
	WHERE table_schema = 'public' AND table_name = $1
)`, name).Scan(&exists)
	return exists, err
}

func columnExists(ctx context.Context, table, column string) (bool, error) {
	pool := postgres.Pool()
	if pool == nil {
		return false, errors.New("postgres pool is not initialized")
	}
	var exists bool
	err := pool.QueryRow(ctx, `
SELECT EXISTS (
	SELECT 1 FROM information_schema.columns
	WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
)`, table, column).Scan(&exists)
	return exists, err
}

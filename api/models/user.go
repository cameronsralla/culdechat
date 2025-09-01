package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// User represents the users table.
type User struct {
	ID                uuid.UUID
	UnitNumber        string
	Email             string
	HashedPassword    string
	ProfilePictureURL *string
	IsDirectoryOptIn  bool
	IsAdmin           bool
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// EnsureUsersTable creates the users table if it doesn't exist.
func EnsureUsersTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    unit_number VARCHAR NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    hashed_password VARCHAR NOT NULL,
    profile_picture_url VARCHAR NULL,
    is_directory_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := pool.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure users table: %v", err)
		return err
	}
	return nil
}

// InsertUser inserts a new user. Caller must provide a hashed password.
func InsertUser(ctx context.Context, u *User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	const q = `
INSERT INTO users (
    id, unit_number, email, hashed_password, profile_picture_url,
    is_directory_opt_in, is_admin, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING created_at, updated_at;
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	return pool.QueryRow(ctx, q,
		u.ID, u.UnitNumber, u.Email, u.HashedPassword, u.ProfilePictureURL,
		u.IsDirectoryOptIn, u.IsAdmin, u.Status,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
}

// GetUserByEmail fetches a user by email.
func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
SELECT id, unit_number, email, hashed_password, profile_picture_url,
       is_directory_opt_in, is_admin, status, created_at, updated_at
FROM users WHERE email = $1 LIMIT 1;
`
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var u User
	var profileURL *string
	err := pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.UnitNumber, &u.Email, &u.HashedPassword, &profileURL,
		&u.IsDirectoryOptIn, &u.IsAdmin, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.ProfilePictureURL = profileURL
	return &u, nil
}

// GetUserByID fetches a user by ID.
func GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const q = `
SELECT id, unit_number, email, hashed_password, profile_picture_url,
       is_directory_opt_in, is_admin, status, created_at, updated_at
FROM users WHERE id = $1 LIMIT 1;
`
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var u User
	var profileURL *string
	err := pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.UnitNumber, &u.Email, &u.HashedPassword, &profileURL,
		&u.IsDirectoryOptIn, &u.IsAdmin, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.ProfilePictureURL = profileURL
	return &u, nil
}

// UpdateUser updates mutable fields and bumps updated_at.
func UpdateUser(ctx context.Context, u *User) error {
	const q = `
UPDATE users SET
    unit_number = $2,
    email = $3,
    hashed_password = $4,
    profile_picture_url = $5,
    is_directory_opt_in = $6,
    is_admin = $7,
    status = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING created_at, updated_at;
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	var createdAt time.Time
	return pool.QueryRow(ctx, q,
		u.ID, u.UnitNumber, u.Email, u.HashedPassword, u.ProfilePictureURL,
		u.IsDirectoryOptIn, u.IsAdmin, u.Status,
	).Scan(&createdAt, &u.UpdatedAt)
}

// SoftDeleteUser flags a user as inactive. Hard delete is handled by retention jobs.
func SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	const q = `
UPDATE users SET status = 'inactive', updated_at = NOW() WHERE id = $1;
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := pool.Exec(ctx, q, id)
	return err
}

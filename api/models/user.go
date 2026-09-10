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

const (
	UserStatusActive   = "active"
	UserStatusPending  = "pending"
	UserStatusInactive = "inactive"
)

// User represents the users table.
type User struct {
	ID                uuid.UUID
	UnitNumber        string
	Email             string
	Name              string
	HashedPassword    string
	ProfilePictureURL *string
	IsDirectoryOptIn  bool
	IsAdmin           bool
	Status            string
	InviteToken       *string // stored as SHA-256 hex of the invite token
	PasscodeHash      *string
	InviteExpiresAt   *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

const userSelectColumns = `
id, unit_number, email, name, hashed_password, profile_picture_url,
is_directory_opt_in, is_admin, status, invite_token, passcode_hash,
invite_expires_at, created_at, updated_at`

func scanUser(row pgx.Row) (*User, error) {
	var u User
	var profileURL *string
	var inviteToken *string
	var passcodeHash *string
	var inviteExpires *time.Time
	err := row.Scan(
		&u.ID, &u.UnitNumber, &u.Email, &u.Name, &u.HashedPassword, &profileURL,
		&u.IsDirectoryOptIn, &u.IsAdmin, &u.Status, &inviteToken, &passcodeHash,
		&inviteExpires, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.ProfilePictureURL = profileURL
	u.InviteToken = inviteToken
	u.PasscodeHash = passcodeHash
	u.InviteExpiresAt = inviteExpires
	return &u, nil
}

// EnsureUsersTable creates the users table if it doesn't exist and applies additive changes.
func EnsureUsersTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    unit_number VARCHAR NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    name VARCHAR NOT NULL DEFAULT '',
    hashed_password VARCHAR NOT NULL,
    profile_picture_url VARCHAR NULL,
    is_directory_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR NOT NULL DEFAULT 'active',
    invite_token VARCHAR NULL,
    passcode_hash VARCHAR NULL,
    invite_expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_invite_token ON users (invite_token) WHERE invite_token IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_unit_active ON users (unit_number) WHERE status IN ('active', 'pending');
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := pool.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure users table: %v", err)
		return err
	}

	alters := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_token VARCHAR NULL`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS passcode_hash VARCHAR NULL`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_expires_at TIMESTAMPTZ NULL`,
	}
	for _, q := range alters {
		if _, err := pool.Exec(ctx, q); err != nil {
			utils.Errorf("failed to alter users table: %v", err)
			return err
		}
	}
	return nil
}

// InsertUser inserts a new user. Caller must provide a hashed password.
func InsertUser(ctx context.Context, u *User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	const q = `
INSERT INTO users (
    id, unit_number, email, name, hashed_password, profile_picture_url,
    is_directory_opt_in, is_admin, status, invite_token, passcode_hash, invite_expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING created_at, updated_at;
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	return pool.QueryRow(ctx, q,
		u.ID, u.UnitNumber, u.Email, u.Name, u.HashedPassword, u.ProfilePictureURL,
		u.IsDirectoryOptIn, u.IsAdmin, u.Status, u.InviteToken, u.PasscodeHash, u.InviteExpiresAt,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
}

// GetUserByEmail fetches a user by email.
func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	q := `SELECT ` + userSelectColumns + ` FROM users WHERE email = $1 LIMIT 1;`
	return scanUser(pool.QueryRow(ctx, q, email))
}

// GetUserByID fetches a user by ID.
func GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	q := `SELECT ` + userSelectColumns + ` FROM users WHERE id = $1 LIMIT 1;`
	return scanUser(pool.QueryRow(ctx, q, id))
}

// GetUserByInviteTokenHash fetches a pending invite by hashed registration token.
func GetUserByInviteTokenHash(ctx context.Context, tokenHash string) (*User, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	q := `SELECT ` + userSelectColumns + ` FROM users WHERE invite_token = $1 LIMIT 1;`
	return scanUser(pool.QueryRow(ctx, q, tokenHash))
}

// UpdateUser updates mutable fields and bumps updated_at.
func UpdateUser(ctx context.Context, u *User) error {
	const q = `
UPDATE users SET
    unit_number = $2,
    email = $3,
    name = $4,
    hashed_password = $5,
    profile_picture_url = $6,
    is_directory_opt_in = $7,
    is_admin = $8,
    status = $9,
    invite_token = $10,
    passcode_hash = $11,
    invite_expires_at = $12,
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
		u.ID, u.UnitNumber, u.Email, u.Name, u.HashedPassword, u.ProfilePictureURL,
		u.IsDirectoryOptIn, u.IsAdmin, u.Status, u.InviteToken, u.PasscodeHash, u.InviteExpiresAt,
	).Scan(&createdAt, &u.UpdatedAt)
}

// SoftDeleteUser flags a user as inactive and clears invite material so the unit can be reused.
func SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	const q = `
UPDATE users SET
	status = 'inactive',
	invite_token = NULL,
	passcode_hash = NULL,
	invite_expires_at = NULL,
	updated_at = NOW()
WHERE id = $1;
`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := pool.Exec(ctx, q, id)
	return err
}

// CountAdmins returns the number of active admin users.
func CountAdmins(ctx context.Context) (int, error) {
	pool := postgres.Pool()
	if pool == nil {
		return 0, errors.New("postgres pool is not initialized")
	}
	var n int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_admin = TRUE AND status = 'active'`).Scan(&n)
	return n, err
}

// AdminUserListItem is a roster row for business admins (no secrets).
type AdminUserListItem struct {
	ID         uuid.UUID
	Email      string
	Name       string
	UnitNumber string
	Status     string
	IsAdmin    bool
}

// ListUsers returns all users, newest first, for admin roster screens.
func ListUsers(ctx context.Context) ([]AdminUserListItem, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	const q = `
SELECT id, email, name, unit_number, status, is_admin
FROM users
ORDER BY created_at DESC;
`
	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AdminUserListItem
	for rows.Next() {
		var u AdminUserListItem
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.UnitNumber, &u.Status, &u.IsAdmin); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

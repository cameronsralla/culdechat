package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// EnsureRefreshTokensTable stores hashed refresh tokens so sessions can be revoked.
func EnsureRefreshTokensTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS refresh_tokens (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	token_hash VARCHAR NOT NULL UNIQUE,
	expires_at TIMESTAMPTZ NOT NULL,
	revoked_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens (user_id);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure refresh_tokens table: %v", err)
		return err
	}
	return nil
}

// InsertRefreshToken stores a hashed refresh token.
func InsertRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, uuid.New(), userID, tokenHash, expiresAt)
	return err
}

// LookupRefreshToken returns user id if the hash is valid, unrevoked, and unexpired.
func LookupRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	const q = `
SELECT user_id FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
LIMIT 1;
`
	p := postgres.Pool()
	if p == nil {
		return uuid.Nil, errors.New("postgres pool is not initialized")
	}
	var id uuid.UUID
	err := p.QueryRow(ctx, q, tokenHash).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// RevokeRefreshToken marks a single token revoked.
func RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL;`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, tokenHash)
	return err
}

// RevokeRefreshTokensForUser revokes every session for a user (logout-all, offboard, password change).
func RevokeRefreshTokensForUser(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL;`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, userID)
	return err
}

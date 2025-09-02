package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// Board represents the boards table.
type Board struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EnsureBoardsTable creates the boards table if it doesn't exist.
func EnsureBoardsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS boards (
	id UUID PRIMARY KEY,
	name VARCHAR NOT NULL UNIQUE,
	description VARCHAR NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_boards_name ON boards (name);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure boards table: %v", err)
		return err
	}
	return nil
}

// InsertBoard inserts a new board.
func InsertBoard(ctx context.Context, b *Board) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	const q = `
INSERT INTO boards (id, name, description)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, b.ID, b.Name, b.Description).Scan(&b.CreatedAt, &b.UpdatedAt)
}

// ListBoards returns all boards, newest first.
func ListBoards(ctx context.Context) ([]Board, error) {
	const q = `
SELECT id, name, description, created_at, updated_at
FROM boards
ORDER BY created_at DESC;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Board
	for rows.Next() {
		var b Board
		var desc *string
		if err := rows.Scan(&b.ID, &b.Name, &desc, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		b.Description = desc
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBoardByID fetches a board by id.
func GetBoardByID(ctx context.Context, id uuid.UUID) (*Board, error) {
	const q = `
SELECT id, name, description, created_at, updated_at
FROM boards WHERE id = $1 LIMIT 1;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var b Board
	var desc *string
	if err := p.QueryRow(ctx, q, id).Scan(&b.ID, &b.Name, &desc, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return nil, err
	}
	b.Description = desc
	return &b, nil
}

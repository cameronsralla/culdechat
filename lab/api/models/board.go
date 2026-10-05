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

type boardRowScanner interface {
	Scan(dest ...any) error
}

// Board represents the boards table.
type Board struct {
	ID          uuid.UUID
	CreatorID   uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BoardListItem is a board plus aggregate fields for listing.
type BoardListItem struct {
	Board
	SubscriberCount int64
	IsSubscribed    bool
}

// EnsureBoardsTable creates the boards table if it doesn't exist.
func EnsureBoardsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS boards (
	id UUID PRIMARY KEY,
	creator_id UUID NULL,
	name VARCHAR NOT NULL UNIQUE,
	description VARCHAR NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_boards_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL
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
	if _, err := p.Exec(ctx, `ALTER TABLE boards ADD COLUMN IF NOT EXISTS creator_id UUID NULL`); err != nil {
		utils.Errorf("failed to alter boards table: %v", err)
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
INSERT INTO boards (id, creator_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, b.ID, b.CreatorID, b.Name, b.Description).Scan(&b.CreatedAt, &b.UpdatedAt)
}

// ListBoardsForUser returns all boards with subscriber counts and whether the given user is subscribed.
func ListBoardsForUser(ctx context.Context, userID uuid.UUID) ([]BoardListItem, error) {
	const q = `
SELECT b.id, b.creator_id, b.name, b.description, b.created_at, b.updated_at,
	COUNT(s.user_id) AS subscriber_count,
	EXISTS (
		SELECT 1 FROM board_subscriptions me
		WHERE me.board_id = b.id AND me.user_id = $1
	) AS is_subscribed
FROM boards b
LEFT JOIN board_subscriptions s ON s.board_id = b.id
GROUP BY b.id
ORDER BY b.created_at DESC;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BoardListItem
	for rows.Next() {
		item, err := scanBoardListItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanBoardListItem(row boardRowScanner) (BoardListItem, error) {
	var item BoardListItem
	var creatorID *uuid.UUID
	var desc *string
	err := row.Scan(
		&item.ID, &creatorID, &item.Name, &desc, &item.CreatedAt, &item.UpdatedAt,
		&item.SubscriberCount, &item.IsSubscribed,
	)
	if err != nil {
		return BoardListItem{}, err
	}
	if creatorID != nil {
		item.CreatorID = *creatorID
	}
	item.Description = desc
	return item, nil
}

// GetBoardByID fetches a board by id.
func GetBoardByID(ctx context.Context, id uuid.UUID) (*Board, error) {
	const q = `
SELECT id, creator_id, name, description, created_at, updated_at
FROM boards WHERE id = $1 LIMIT 1;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var b Board
	var creatorID *uuid.UUID
	var desc *string
	err := p.QueryRow(ctx, q, id).Scan(&b.ID, &creatorID, &b.Name, &desc, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if creatorID != nil {
		b.CreatorID = *creatorID
	}
	b.Description = desc
	return &b, nil
}

// GetBoardForUser returns a board with subscriber aggregates for the current user.
func GetBoardForUser(ctx context.Context, userID, boardID uuid.UUID) (*BoardListItem, error) {
	const q = `
SELECT b.id, b.creator_id, b.name, b.description, b.created_at, b.updated_at,
	COUNT(s.user_id) AS subscriber_count,
	EXISTS (
		SELECT 1 FROM board_subscriptions me
		WHERE me.board_id = b.id AND me.user_id = $1
	) AS is_subscribed
FROM boards b
LEFT JOIN board_subscriptions s ON s.board_id = b.id
WHERE b.id = $2
GROUP BY b.id;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	item, err := scanBoardListItem(p.QueryRow(ctx, q, userID, boardID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

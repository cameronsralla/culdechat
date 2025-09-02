package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// Post represents a post/thread in a board. Bulletin posts are marked with IsBulletin=true.
type Post struct {
	ID         uuid.UUID
	BoardID    uuid.UUID
	AuthorID   uuid.UUID
	Title      string
	Content    string
	IsBulletin bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// EnsurePostsTable creates the posts table if it doesn't exist.
func EnsurePostsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS posts (
	id UUID PRIMARY KEY,
	board_id UUID NOT NULL,
	author_id UUID NOT NULL,
	title VARCHAR NOT NULL,
	content TEXT NOT NULL,
	is_bulletin BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_posts_board FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
	CONSTRAINT fk_posts_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_board ON posts (board_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_author ON posts (author_id);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure posts table: %v", err)
		return err
	}
	return nil
}

// InsertPost inserts a new post.
func InsertPost(ctx context.Context, pst *Post) error {
	if pst.ID == uuid.Nil {
		pst.ID = uuid.New()
	}
	const q = `
INSERT INTO posts (id, board_id, author_id, title, content, is_bulletin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, pst.ID, pst.BoardID, pst.AuthorID, pst.Title, pst.Content, pst.IsBulletin).Scan(&pst.CreatedAt, &pst.UpdatedAt)
}

// ListPostsByBoard returns posts for a board, newest first.
func ListPostsByBoard(ctx context.Context, boardID uuid.UUID) ([]Post, error) {
	const q = `
SELECT id, board_id, author_id, title, content, is_bulletin, created_at, updated_at
FROM posts WHERE board_id = $1
ORDER BY created_at DESC;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Post
	for rows.Next() {
		var pst Post
		if err := rows.Scan(&pst.ID, &pst.BoardID, &pst.AuthorID, &pst.Title, &pst.Content, &pst.IsBulletin, &pst.CreatedAt, &pst.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, pst)
	}
	return out, rows.Err()
}

// ListBulletins returns bulletin posts, newest first.
func ListBulletins(ctx context.Context) ([]Post, error) {
	const q = `
SELECT id, board_id, author_id, title, content, is_bulletin, created_at, updated_at
FROM posts WHERE is_bulletin = TRUE
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
	var out []Post
	for rows.Next() {
		var pst Post
		if err := rows.Scan(&pst.ID, &pst.BoardID, &pst.AuthorID, &pst.Title, &pst.Content, &pst.IsBulletin, &pst.CreatedAt, &pst.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, pst)
	}
	return out, rows.Err()
}

// GetPostByID fetches a single post by id.
func GetPostByID(ctx context.Context, id uuid.UUID) (*Post, error) {
	const q = `
SELECT id, board_id, author_id, title, content, is_bulletin, created_at, updated_at
FROM posts WHERE id = $1 LIMIT 1;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var pst Post
	if err := p.QueryRow(ctx, q, id).Scan(&pst.ID, &pst.BoardID, &pst.AuthorID, &pst.Title, &pst.Content, &pst.IsBulletin, &pst.CreatedAt, &pst.UpdatedAt); err != nil {
		return nil, err
	}
	return &pst, nil
}

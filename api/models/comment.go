package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// Comment represents a comment on a post.
type Comment struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EnsureCommentsTable creates the comments table if it doesn't exist.
func EnsureCommentsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS comments (
	id UUID PRIMARY KEY,
	post_id UUID NOT NULL,
	author_id UUID NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comments_post ON comments (post_id, created_at ASC);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure comments table: %v", err)
		return err
	}
	return nil
}

// InsertComment inserts a new comment.
func InsertComment(ctx context.Context, cmt *Comment) error {
	if cmt.ID == uuid.Nil {
		cmt.ID = uuid.New()
	}
	const q = `
INSERT INTO comments (id, post_id, author_id, content)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, cmt.ID, cmt.PostID, cmt.AuthorID, cmt.Content).Scan(&cmt.CreatedAt, &cmt.UpdatedAt)
}

// ListCommentsByPost returns comments for a post in chronological order.
func ListCommentsByPost(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	const q = `
SELECT id, post_id, author_id, content, created_at, updated_at
FROM comments WHERE post_id = $1
ORDER BY created_at ASC;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Comment
	for rows.Next() {
		var cmt Comment
		if err := rows.Scan(&cmt.ID, &cmt.PostID, &cmt.AuthorID, &cmt.Content, &cmt.CreatedAt, &cmt.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, cmt)
	}
	return out, rows.Err()
}

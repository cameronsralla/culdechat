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

// Comment represents a comment on a post.
type Comment struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CommentWithAuthor is a comment plus public author fields.
type CommentWithAuthor struct {
	Comment
	AuthorUnit string
	AuthorName *string
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

// ListCommentsByPost returns comments for a post in chronological order with author info.
func ListCommentsByPost(ctx context.Context, postID uuid.UUID) ([]CommentWithAuthor, error) {
	const q = `
SELECT c.id, c.post_id, c.author_id, c.content, c.created_at, c.updated_at,
	u.unit_number,
	CASE WHEN u.is_directory_opt_in THEN NULLIF(u.name, '') ELSE NULL END AS author_name
FROM comments c
JOIN users u ON u.id = c.author_id
WHERE c.post_id = $1
ORDER BY c.created_at ASC;
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

	var out []CommentWithAuthor
	for rows.Next() {
		var cmt CommentWithAuthor
		var authorName *string
		if err := rows.Scan(
			&cmt.ID, &cmt.PostID, &cmt.AuthorID, &cmt.Content, &cmt.CreatedAt, &cmt.UpdatedAt,
			&cmt.AuthorUnit, &authorName,
		); err != nil {
			return nil, err
		}
		cmt.AuthorName = authorName
		out = append(out, cmt)
	}
	return out, rows.Err()
}

// GetCommentByID fetches a comment without author join.
func GetCommentByID(ctx context.Context, id uuid.UUID) (*Comment, error) {
	const q = `
SELECT id, post_id, author_id, content, created_at, updated_at
FROM comments WHERE id = $1 LIMIT 1;
`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	var cmt Comment
	err := p.QueryRow(ctx, q, id).Scan(&cmt.ID, &cmt.PostID, &cmt.AuthorID, &cmt.Content, &cmt.CreatedAt, &cmt.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cmt, nil
}

// UpdateCommentContent updates comment text.
func UpdateCommentContent(ctx context.Context, id uuid.UUID, content string) error {
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, `UPDATE comments SET content = $2, updated_at = NOW() WHERE id = $1;`, id, content)
	return err
}

// DeleteComment deletes a comment.
func DeleteComment(ctx context.Context, id uuid.UUID) error {
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, `DELETE FROM comments WHERE id = $1;`, id)
	return err
}

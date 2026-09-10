package models

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rowScanner interface {
	Scan(dest ...any) error
}

const (
	PostTypeStandard = "standard"
	PostTypeBulletin = "bulletin"
)

// Post represents a post/thread in a board.
type Post struct {
	ID        uuid.UUID
	BoardID   uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Content   string
	PostType  string
	IsPinned  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FeedPost is a post plus author, board, and count fields for list/detail views.
type FeedPost struct {
	Post
	AuthorUnit       string
	AuthorName       *string
	BoardName        string
	CommentCount     int64
	ReactionCount    int64
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
	post_type VARCHAR NOT NULL DEFAULT 'standard',
	is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_posts_board FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
	CONSTRAINT fk_posts_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_board ON posts (board_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_feed ON posts (is_pinned DESC, created_at DESC, id DESC);
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
	if _, err := p.Exec(ctx, `ALTER TABLE posts ADD COLUMN IF NOT EXISTS post_type VARCHAR NOT NULL DEFAULT 'standard'`); err != nil {
		return err
	}
	if _, err := p.Exec(ctx, `ALTER TABLE posts ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN NOT NULL DEFAULT FALSE`); err != nil {
		return err
	}

	hasBulletin, err := columnExists(ctx, "posts", "is_bulletin")
	if err != nil {
		return err
	}
	if hasBulletin {
		if _, err := p.Exec(ctx, `
UPDATE posts SET post_type = 'bulletin', is_pinned = TRUE
WHERE is_bulletin = TRUE AND post_type = 'standard'
`); err != nil {
			return err
		}
	}
	return nil
}

// InsertPost inserts a new post.
func InsertPost(ctx context.Context, pst *Post) error {
	if pst.ID == uuid.Nil {
		pst.ID = uuid.New()
	}
	if pst.PostType == "" {
		pst.PostType = PostTypeStandard
	}
	const q = `
INSERT INTO posts (id, board_id, author_id, title, content, post_type, is_pinned)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, pst.ID, pst.BoardID, pst.AuthorID, pst.Title, pst.Content, pst.PostType, pst.IsPinned).Scan(&pst.CreatedAt, &pst.UpdatedAt)
}

const feedSelect = `
SELECT p.id, p.board_id, p.author_id, p.title, p.content, p.post_type, p.is_pinned, p.created_at, p.updated_at,
	u.unit_number,
	CASE WHEN u.is_directory_opt_in THEN NULLIF(u.name, '') ELSE NULL END AS author_name,
	b.name AS board_name,
	(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS comment_count,
	(SELECT COUNT(*) FROM post_reactions r WHERE r.post_id = p.id) AS reaction_count
FROM posts p
JOIN users u ON u.id = p.author_id
JOIN boards b ON b.id = p.board_id
`

func scanFeedPost(row rowScanner) (FeedPost, error) {
	var fp FeedPost
	var authorName *string
	err := row.Scan(
		&fp.ID, &fp.BoardID, &fp.AuthorID, &fp.Title, &fp.Content, &fp.PostType, &fp.IsPinned, &fp.CreatedAt, &fp.UpdatedAt,
		&fp.AuthorUnit, &authorName, &fp.BoardName, &fp.CommentCount, &fp.ReactionCount,
	)
	if err != nil {
		return FeedPost{}, err
	}
	fp.AuthorName = authorName
	return fp, nil
}

// ListFeed returns a page of posts, pinned first, then newest. If boardID is set, results are limited to that board.
func ListFeed(ctx context.Context, boardID *uuid.UUID, limit int, cursorPinned *bool, cursorCreated *time.Time, cursorID *uuid.UUID) ([]FeedPost, error) {
	if limit <= 0 {
		limit = 20
	}

	q := feedSelect + ` WHERE 1=1`
	args := []any{}
	arg := 1

	if boardID != nil {
		q += ` AND p.board_id = $` + strconv.Itoa(arg)
		args = append(args, *boardID)
		arg++
	}

	if cursorPinned != nil && cursorCreated != nil && cursorID != nil {
		// Match ORDER BY is_pinned DESC, created_at DESC, id DESC
		q += ` AND (
			p.is_pinned < $` + strconv.Itoa(arg) + `
			OR (p.is_pinned = $` + strconv.Itoa(arg) + ` AND p.created_at < $` + strconv.Itoa(arg+1) + `)
			OR (p.is_pinned = $` + strconv.Itoa(arg) + ` AND p.created_at = $` + strconv.Itoa(arg+1) + ` AND p.id < $` + strconv.Itoa(arg+2) + `)
		)`
		args = append(args, *cursorPinned, *cursorCreated, *cursorID)
		arg += 3
	}

	q += ` ORDER BY p.is_pinned DESC, p.created_at DESC, p.id DESC LIMIT $` + strconv.Itoa(arg)
	args = append(args, limit)

	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FeedPost
	for rows.Next() {
		fp, err := scanFeedPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, fp)
	}
	return out, rows.Err()
}

// GetFeedPostByID fetches a single post with author/board/counts.
func GetFeedPostByID(ctx context.Context, id uuid.UUID) (*FeedPost, error) {
	q := feedSelect + ` WHERE p.id = $1 LIMIT 1;`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	fp, err := scanFeedPost(p.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &fp, nil
}

// UpdatePost updates title, content, and pin flag.
func UpdatePost(ctx context.Context, pst *Post) error {
	const q = `
UPDATE posts SET title = $2, content = $3, is_pinned = $4, updated_at = NOW()
WHERE id = $1
RETURNING updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, pst.ID, pst.Title, pst.Content, pst.IsPinned).Scan(&pst.UpdatedAt)
}

// DeletePost deletes a post (comments and reactions cascade).
func DeletePost(ctx context.Context, id uuid.UUID) error {
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, `DELETE FROM posts WHERE id = $1;`, id)
	return err
}

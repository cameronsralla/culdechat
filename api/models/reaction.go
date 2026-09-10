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

// Reaction is a user's emoji reaction on a post. One per user per post.
type Reaction struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	UserID    uuid.UUID
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EnsurePostReactionsTable creates the post_reactions table with a uniqueness constraint per user/post.
func EnsurePostReactionsTable(ctx context.Context) error {
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}

	oldExists, err := tableExists(ctx, "reactions")
	if err != nil {
		return err
	}
	newExists, err := tableExists(ctx, "post_reactions")
	if err != nil {
		return err
	}
	if oldExists && !newExists {
		if _, err := p.Exec(ctx, `ALTER TABLE reactions RENAME TO post_reactions`); err != nil {
			utils.Errorf("failed to rename reactions table: %v", err)
			return err
		}
	}

	const ddl = `
CREATE TABLE IF NOT EXISTS post_reactions (
	id UUID PRIMARY KEY,
	post_id UUID NOT NULL,
	user_id UUID NOT NULL,
	type VARCHAR NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_reactions_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	CONSTRAINT fk_reactions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT uq_reaction_user_post UNIQUE (post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_reactions_post ON post_reactions (post_id);
`
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure post_reactions table: %v", err)
		return err
	}
	return nil
}

// UpsertReaction inserts or updates a user's reaction on a post.
func UpsertReaction(ctx context.Context, r *Reaction) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	const q = `
INSERT INTO post_reactions (id, post_id, user_id, type)
VALUES ($1, $2, $3, $4)
ON CONFLICT (post_id, user_id)
DO UPDATE SET type = EXCLUDED.type, updated_at = NOW()
RETURNING created_at, updated_at;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	return p.QueryRow(ctx, q, r.ID, r.PostID, r.UserID, r.Type).Scan(&r.CreatedAt, &r.UpdatedAt)
}

// RemoveReaction deletes a user's reaction from a post.
func RemoveReaction(ctx context.Context, postID, userID uuid.UUID) error {
	const q = `
DELETE FROM post_reactions WHERE post_id = $1 AND user_id = $2;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, postID, userID)
	return err
}

// ReactionCount is a grouped reaction tally.
type ReactionCount struct {
	Type  string
	Count int64
}

// CountReactionsByPost returns reaction counts grouped by type.
func CountReactionsByPost(ctx context.Context, postID uuid.UUID) ([]ReactionCount, error) {
	const q = `
SELECT type, COUNT(*)
FROM post_reactions WHERE post_id = $1
GROUP BY type;
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
	var out []ReactionCount
	for rows.Next() {
		var rc ReactionCount
		if err := rows.Scan(&rc.Type, &rc.Count); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}

// GetReactionType returns the current user's reaction type on a post, or empty.
func GetReactionType(ctx context.Context, postID, userID uuid.UUID) (string, error) {
	const q = `SELECT type FROM post_reactions WHERE post_id = $1 AND user_id = $2 LIMIT 1;`
	p := postgres.Pool()
	if p == nil {
		return "", errors.New("postgres pool is not initialized")
	}
	var t string
	err := p.QueryRow(ctx, q, postID, userID).Scan(&t)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return t, nil
}

// GetReactionTypesForPosts returns post_id -> type for one user across many posts.
func GetReactionTypesForPosts(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	out := map[uuid.UUID]string{}
	if len(postIDs) == 0 {
		return out, nil
	}
	const q = `SELECT post_id, type FROM post_reactions WHERE user_id = $1 AND post_id = ANY($2);`
	p := postgres.Pool()
	if p == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := p.Query(ctx, q, userID, postIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var t string
		if err := rows.Scan(&id, &t); err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, rows.Err()
}

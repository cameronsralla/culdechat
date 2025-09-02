package models

import (
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// ReactionType is a constrained set of supported emoji keywords for MVP.
// Stored as a short string, e.g., "like", "love", "laugh", "wow", "sad", "angry".
type Reaction struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	UserID    uuid.UUID
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EnsureReactionsTable creates the reactions table with a uniqueness constraint per user/post.
func EnsureReactionsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS reactions (
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

CREATE INDEX IF NOT EXISTS idx_reactions_post ON reactions (post_id);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure reactions table: %v", err)
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
INSERT INTO reactions (id, post_id, user_id, type)
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
DELETE FROM reactions WHERE post_id = $1 AND user_id = $2;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, postID, userID)
	return err
}

// CountReactionsByPost returns reaction counts grouped by type.
type ReactionCount struct {
	Type  string
	Count int64
}

func CountReactionsByPost(ctx context.Context, postID uuid.UUID) ([]ReactionCount, error) {
	const q = `
SELECT type, COUNT(*)
FROM reactions WHERE post_id = $1
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

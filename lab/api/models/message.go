package models

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Conversation is a unique 1:1 thread between two users.
// user_low_id / user_high_id are ordered so the pair is unique regardless of who started it.
type Conversation struct {
	ID         uuid.UUID
	UserLowID  uuid.UUID
	UserHighID uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Message is one text message in a conversation.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
	CreatedAt      time.Time
}

// ConversationListItem is an inbox row for the current user.
type ConversationListItem struct {
	Conversation
	PeerID                uuid.UUID
	PeerUnit              string
	PeerName              *string
	PeerProfilePictureURL *string
	PeerDirectoryOptIn    bool
	LastMessageID         *uuid.UUID
	LastMessageContent    *string
	LastMessageSenderID   *uuid.UUID
	LastMessageAt         *time.Time
}

// OrderedPair returns (low, high) user IDs matching Postgres UUID ordering.
func OrderedPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) <= 0 {
		return a, b
	}
	return b, a
}

// EnsureConversationsTable creates the conversations table.
func EnsureConversationsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS conversations (
	id UUID PRIMARY KEY,
	user_low_id UUID NOT NULL,
	user_high_id UUID NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_conversations_low FOREIGN KEY (user_low_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_conversations_high FOREIGN KEY (user_high_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT chk_conversations_ordered CHECK (user_low_id < user_high_id),
	CONSTRAINT uq_conversations_pair UNIQUE (user_low_id, user_high_id)
);

CREATE INDEX IF NOT EXISTS idx_conversations_low ON conversations (user_low_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_high ON conversations (user_high_id, updated_at DESC);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure conversations table: %v", err)
		return err
	}
	return nil
}

// EnsureMessagesTable creates the direct messages table.
func EnsureMessagesTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS messages (
	id UUID PRIMARY KEY,
	conversation_id UUID NOT NULL,
	sender_id UUID NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_messages_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
	CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages (conversation_id, created_at ASC);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure messages table: %v", err)
		return err
	}
	return nil
}

// GetActiveUserByUnit returns the active resident for a unit number (today: one per unit).
func GetActiveUserByUnit(ctx context.Context, unitNumber string) (*User, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	q := `SELECT ` + userSelectColumns + `
FROM users
WHERE unit_number = $1 AND status = 'active'
LIMIT 1;`
	return scanUser(pool.QueryRow(ctx, q, unitNumber))
}

// GetConversationByID loads a conversation by id.
func GetConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	const q = `
SELECT id, user_low_id, user_high_id, created_at, updated_at
FROM conversations WHERE id = $1 LIMIT 1;`
	var c Conversation
	err := pool.QueryRow(ctx, q, id).Scan(&c.ID, &c.UserLowID, &c.UserHighID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// GetConversationByPair finds the conversation for two users, if any.
func GetConversationByPair(ctx context.Context, a, b uuid.UUID) (*Conversation, error) {
	low, high := OrderedPair(a, b)
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	const q = `
SELECT id, user_low_id, user_high_id, created_at, updated_at
FROM conversations WHERE user_low_id = $1 AND user_high_id = $2 LIMIT 1;`
	var c Conversation
	err := pool.QueryRow(ctx, q, low, high).Scan(&c.ID, &c.UserLowID, &c.UserHighID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// InsertConversation creates a new 1:1 conversation.
func InsertConversation(ctx context.Context, a, b uuid.UUID) (*Conversation, error) {
	low, high := OrderedPair(a, b)
	c := &Conversation{
		ID:         uuid.New(),
		UserLowID:  low,
		UserHighID: high,
	}
	const q = `
INSERT INTO conversations (id, user_low_id, user_high_id)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at;`
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	if err := pool.QueryRow(ctx, q, c.ID, c.UserLowID, c.UserHighID).Scan(&c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return c, nil
}

// TouchConversation bumps updated_at (after a new message).
func TouchConversation(ctx context.Context, id uuid.UUID) error {
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := pool.Exec(ctx, `UPDATE conversations SET updated_at = NOW() WHERE id = $1;`, id)
	return err
}

// InsertMessage inserts a message row.
func InsertMessage(ctx context.Context, m *Message) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	const q = `
INSERT INTO messages (id, conversation_id, sender_id, content)
VALUES ($1, $2, $3, $4)
RETURNING created_at;`
	pool := postgres.Pool()
	if pool == nil {
		return errors.New("postgres pool is not initialized")
	}
	return pool.QueryRow(ctx, q, m.ID, m.ConversationID, m.SenderID, m.Content).Scan(&m.CreatedAt)
}

// ListConversationsForUser returns inbox rows newest-activity first.
func ListConversationsForUser(ctx context.Context, userID uuid.UUID) ([]ConversationListItem, error) {
	const q = `
SELECT
	c.id, c.user_low_id, c.user_high_id, c.created_at, c.updated_at,
	peer.id,
	peer.unit_number,
	CASE WHEN peer.is_directory_opt_in THEN NULLIF(peer.name, '') ELSE NULL END,
	CASE WHEN peer.is_directory_opt_in THEN peer.profile_picture_url ELSE NULL END,
	peer.is_directory_opt_in,
	lm.id, lm.content, lm.sender_id, lm.created_at
FROM conversations c
JOIN users peer ON peer.id = CASE
	WHEN c.user_low_id = $1 THEN c.user_high_id
	ELSE c.user_low_id
END
LEFT JOIN LATERAL (
	SELECT id, content, sender_id, created_at
	FROM messages
	WHERE conversation_id = c.id
	ORDER BY created_at DESC
	LIMIT 1
) lm ON TRUE
WHERE c.user_low_id = $1 OR c.user_high_id = $1
ORDER BY c.updated_at DESC;
`
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	rows, err := pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ConversationListItem
	for rows.Next() {
		var item ConversationListItem
		if err := rows.Scan(
			&item.ID, &item.UserLowID, &item.UserHighID, &item.CreatedAt, &item.UpdatedAt,
			&item.PeerID, &item.PeerUnit, &item.PeerName, &item.PeerProfilePictureURL, &item.PeerDirectoryOptIn,
			&item.LastMessageID, &item.LastMessageContent, &item.LastMessageSenderID, &item.LastMessageAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListMessages returns messages in a conversation, oldest first, with optional before-cursor pagination.
// When beforeID is set, returns messages strictly older than that message (for loading history).
func ListMessages(ctx context.Context, conversationID uuid.UUID, limit int, beforeID *uuid.UUID) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}

	var rows pgx.Rows
	var err error
	if beforeID != nil {
		const q = `
SELECT id, conversation_id, sender_id, content, created_at
FROM messages
WHERE conversation_id = $1
  AND created_at < (SELECT created_at FROM messages WHERE id = $2 AND conversation_id = $1)
ORDER BY created_at DESC
LIMIT $3;`
		rows, err = pool.Query(ctx, q, conversationID, *beforeID, limit)
	} else {
		const q = `
SELECT id, conversation_id, sender_id, content, created_at
FROM (
	SELECT id, conversation_id, sender_id, content, created_at
	FROM messages
	WHERE conversation_id = $1
	ORDER BY created_at DESC
	LIMIT $2
) recent
ORDER BY created_at ASC;`
		rows, err = pool.Query(ctx, q, conversationID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if beforeID != nil {
		// Queried DESC; reverse to ASC for the client.
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return out, nil
}

// MessageRecipient is a search hit for starting a DM.
type MessageRecipient struct {
	Kind              string // "user" or "unit"
	UserID            *string
	Name              *string
	UnitNumber        string
	ProfilePictureURL *string
}

// SearchMessageRecipients finds visible users by name/unit and hidden residents by unit only.
// Excludes the requester. Hidden hits never expose name, photo, or user id.
func SearchMessageRecipients(ctx context.Context, requesterID uuid.UUID, query string, limit int) ([]MessageRecipient, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	pool := postgres.Pool()
	if pool == nil {
		return nil, errors.New("postgres pool is not initialized")
	}
	pattern := "%" + query + "%"
	const q = `
(
	SELECT
		'user'::text AS kind,
		id::text AS user_id,
		NULLIF(name, '') AS name,
		unit_number,
		profile_picture_url
	FROM users
	WHERE status = 'active'
	  AND id <> $1
	  AND is_directory_opt_in = TRUE
	  AND (
		unit_number ILIKE $2
		OR name ILIKE $2
	  )
)
UNION ALL
(
	SELECT
		'unit'::text AS kind,
		NULL::text AS user_id,
		NULL::text AS name,
		unit_number,
		NULL::text AS profile_picture_url
	FROM users
	WHERE status = 'active'
	  AND id <> $1
	  AND is_directory_opt_in = FALSE
	  AND unit_number ILIKE $2
)
ORDER BY unit_number ASC
LIMIT $3;`
	rows, err := pool.Query(ctx, q, requesterID, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MessageRecipient
	for rows.Next() {
		var r MessageRecipient
		if err := rows.Scan(&r.Kind, &r.UserID, &r.Name, &r.UnitNumber, &r.ProfilePictureURL); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

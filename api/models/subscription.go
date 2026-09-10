package models

import (
	"context"
	"errors"

	"github.com/cameronsralla/culdechat/connectors/postgres"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/google/uuid"
)

// EnsureBoardSubscriptionsTable creates the board_subscriptions junction table.
func EnsureBoardSubscriptionsTable(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS board_subscriptions (
	user_id UUID NOT NULL,
	board_id UUID NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, board_id),
	CONSTRAINT fk_subs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_subs_board FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	if _, err := p.Exec(ctx, ddl); err != nil {
		utils.Errorf("failed to ensure board_subscriptions table: %v", err)
		return err
	}
	return nil
}

// IsSubscribed reports whether the user is subscribed to the board.
func IsSubscribed(ctx context.Context, userID, boardID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM board_subscriptions WHERE user_id = $1 AND board_id = $2);`
	p := postgres.Pool()
	if p == nil {
		return false, errors.New("postgres pool is not initialized")
	}
	var ok bool
	if err := p.QueryRow(ctx, q, userID, boardID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

// Subscribe adds a subscription. It is idempotent.
func Subscribe(ctx context.Context, userID, boardID uuid.UUID) error {
	const q = `
INSERT INTO board_subscriptions (user_id, board_id)
VALUES ($1, $2)
ON CONFLICT (user_id, board_id) DO NOTHING;
`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, userID, boardID)
	return err
}

// Unsubscribe removes a subscription. It is idempotent.
func Unsubscribe(ctx context.Context, userID, boardID uuid.UUID) error {
	const q = `DELETE FROM board_subscriptions WHERE user_id = $1 AND board_id = $2;`
	p := postgres.Pool()
	if p == nil {
		return errors.New("postgres pool is not initialized")
	}
	_, err := p.Exec(ctx, q, userID, boardID)
	return err
}

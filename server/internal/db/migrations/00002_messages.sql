-- +goose Up
CREATE TABLE conversations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_low     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_high    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL CHECK (status IN ('pending', 'open', 'declined')),
    requested_by UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_low, user_high),
    CHECK (user_low < user_high)
);
CREATE INDEX conversations_low_idx ON conversations (user_low);
CREATE INDEX conversations_high_idx ON conversations (user_high);

CREATE TABLE direct_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id),
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX direct_messages_conversation_idx ON direct_messages (conversation_id, created_at);

-- +goose Down
DROP TABLE direct_messages;
DROP TABLE conversations;

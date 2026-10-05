-- name: FindConversation :one
SELECT id, user_low, user_high, status, requested_by, created_at, updated_at
FROM conversations
WHERE user_low = $1 AND user_high = $2;

-- name: GetConversation :one
SELECT id, user_low, user_high, status, requested_by, created_at, updated_at
FROM conversations
WHERE id = $1;

-- name: InsertConversation :one
INSERT INTO conversations (user_low, user_high, status, requested_by)
VALUES ($1, $2, $3, $4)
RETURNING id, user_low, user_high, status, requested_by, created_at, updated_at;

-- name: SetConversationState :one
UPDATE conversations
SET status = $2, requested_by = $3, updated_at = now()
WHERE id = $1
RETURNING id, user_low, user_high, status, requested_by, created_at, updated_at;

-- name: TouchConversation :exec
UPDATE conversations SET updated_at = now() WHERE id = $1;

-- name: InsertDirectMessage :one
INSERT INTO direct_messages (conversation_id, sender_id, body)
VALUES ($1, $2, $3)
RETURNING id, conversation_id, sender_id, body, created_at;

-- name: ListDirectMessages :many
SELECT id, conversation_id, sender_id, body, created_at
FROM direct_messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: ListConversationsForUser :many
SELECT
    c.id, c.user_low, c.user_high, c.status, c.requested_by, c.created_at, c.updated_at,
    m.body AS last_message
FROM conversations c
JOIN LATERAL (
    SELECT body
    FROM direct_messages
    WHERE conversation_id = c.id
    ORDER BY created_at DESC
    LIMIT 1
) m ON TRUE
WHERE (c.user_low = $1 OR c.user_high = $1)
  AND (c.status <> 'declined' OR c.requested_by = $1)
ORDER BY c.updated_at DESC;

-- name: GetUserContact :one
SELECT id, unit_number, display_name, directory_opt_in, status
FROM residents
WHERE id = $1;

-- name: ActiveHiddenByUnit :one
SELECT id
FROM residents
WHERE status = 'active' AND directory_opt_in = FALSE AND is_primary = TRUE AND unit_number = $1
LIMIT 1;

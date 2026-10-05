-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY unit_number, display_name;

-- name: ListDirectory :many
SELECT id, unit_number, display_name, email
FROM users
WHERE status = 'active' AND directory_opt_in = TRUE
ORDER BY unit_number, display_name;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: CreateInvitedUser :one
INSERT INTO users (email, unit_number, display_name, is_admin, status)
VALUES ($1, $2, $3, $4, 'invited')
RETURNING *;

-- name: CreateActiveUser :one
INSERT INTO users (email, unit_number, display_name, password_hash, is_admin, status)
VALUES ($1, $2, $3, $4, $5, 'active')
RETURNING *;

-- name: ActivateUser :one
UPDATE users
SET password_hash = $2, display_name = $3, status = 'active', updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = $2, directory_opt_in = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetUserAdmin :one
UPDATE users SET is_admin = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: SetUserStatus :one
UPDATE users
SET status = $2,
    deactivated_at = CASE WHEN $2 = 'inactive' THEN now() ELSE NULL END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountActiveAdmins :one
SELECT count(*) FROM users WHERE is_admin = TRUE AND status = 'active';

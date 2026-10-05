-- name: GetUserByID :one
SELECT * FROM residents WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM residents WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM residents ORDER BY unit_number, display_name;

-- name: ListDirectory :many
SELECT id, unit_number, display_name, email
FROM residents
WHERE status = 'active' AND directory_opt_in = TRUE AND is_primary = TRUE
ORDER BY unit_number, display_name;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: CreateUnit :one
INSERT INTO units (number) VALUES ($1) RETURNING *;

-- name: GetUnit :one
SELECT * FROM units WHERE id = $1;

-- name: UpdateUnitNumber :one
UPDATE units SET number = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: DeleteUnit :exec
DELETE FROM units WHERE id = $1;

-- name: CountUnitResidents :one
SELECT count(*) FROM users WHERE unit_id = $1;

-- name: CreateInvitedUser :one
INSERT INTO users (email, unit_id, is_primary, display_name, is_admin, status)
VALUES ($1, $2, TRUE, $3, $4, 'invited')
RETURNING id;

-- name: CreateActiveUser :one
INSERT INTO users (email, unit_id, is_primary, display_name, password_hash, is_admin, status)
VALUES ($1, $2, TRUE, $3, $4, $5, 'active')
RETURNING id;

-- name: ActivateUser :exec
UPDATE users
SET password_hash = $2, display_name = $3, status = 'active', updated_at = now()
WHERE users.id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = now() WHERE users.id = $1;

-- name: UpdateUserProfile :exec
UPDATE users
SET display_name = $2, directory_opt_in = $3, updated_at = now()
WHERE users.id = $1;

-- name: SetUserAdmin :exec
UPDATE users SET is_admin = $2, updated_at = now() WHERE users.id = $1;

-- name: SetUserStatus :exec
UPDATE users
SET status = $2,
    deactivated_at = CASE WHEN $2 = 'inactive' THEN now() ELSE NULL END,
    updated_at = now()
WHERE users.id = $1;

-- name: CountActiveAdmins :one
SELECT count(*) FROM users WHERE is_admin = TRUE AND status = 'active';

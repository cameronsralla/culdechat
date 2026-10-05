-- +goose Up
CREATE TABLE units (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number     TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO units (number)
SELECT DISTINCT unit_number FROM users;

ALTER TABLE users ADD COLUMN unit_id UUID REFERENCES units(id);
ALTER TABLE users ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users AS u
SET unit_id = un.id
FROM units AS un
WHERE un.number = u.unit_number;

WITH ranked AS (
    SELECT id, row_number() OVER (PARTITION BY unit_id ORDER BY created_at ASC, id ASC) AS n
    FROM users
)
UPDATE users AS u
SET is_primary = TRUE
FROM ranked
WHERE u.id = ranked.id AND ranked.n = 1;

ALTER TABLE users ALTER COLUMN unit_id SET NOT NULL;
DROP INDEX users_unit_idx;
ALTER TABLE users DROP COLUMN unit_number;
CREATE INDEX users_unit_idx ON users (unit_id);
CREATE UNIQUE INDEX users_one_primary_per_unit ON users (unit_id) WHERE is_primary;

CREATE VIEW residents AS
SELECT
    u.id,
    u.email,
    un.number AS unit_number,
    u.unit_id,
    u.is_primary,
    u.display_name,
    u.password_hash,
    u.is_admin,
    u.status,
    u.directory_opt_in,
    u.created_at,
    u.updated_at,
    u.deactivated_at
FROM users AS u
JOIN units AS un ON un.id = u.unit_id;

-- +goose Down
DROP VIEW residents;
ALTER TABLE users ADD COLUMN unit_number TEXT;
UPDATE users AS u SET unit_number = un.number FROM units AS un WHERE un.id = u.unit_id;
DROP INDEX users_one_primary_per_unit;
DROP INDEX users_unit_idx;
ALTER TABLE users DROP COLUMN is_primary;
ALTER TABLE users DROP COLUMN unit_id;
ALTER TABLE users ALTER COLUMN unit_number SET NOT NULL;
CREATE INDEX users_unit_idx ON users (unit_number);
DROP TABLE units;

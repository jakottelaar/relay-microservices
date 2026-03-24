-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUsersByIDs :many
SELECT *
FROM users
WHERE id = ANY($1::bigint[]);

-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    avatar,
    bio
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
ON CONFLICT (id) DO UPDATE
SET username = users.username
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET
    avatar = COALESCE(sqlc.narg('avatar'), avatar),
    bio = COALESCE(sqlc.narg('bio'), bio),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
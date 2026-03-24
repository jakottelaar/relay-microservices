-- name: CreateGuild :one
INSERT INTO guilds (
    id,
    name,
    description,
    icon,
    owner_id
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
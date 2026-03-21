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
    COALESCE($4, NULL),
    $5
)
RETURNING *;
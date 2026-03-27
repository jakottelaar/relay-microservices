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

-- name: GetGuild :one
SELECT
    id,
    name,
    description,
    icon,
    owner_id,
    created_at,
    updated_at
FROM guilds
WHERE id = $1;

-- name: CreateGuildChannel :one
INSERT INTO channels (
    id,
    guild_id,
    name,
    topic,
    type
)VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: CreateGuildMember :one
INSERT INTO guild_members (
    guild_id,
    user_id,
    nick
) VALUES (
    $1,
    $2,
    $3
) RETURNING *;
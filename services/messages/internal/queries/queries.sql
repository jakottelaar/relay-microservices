-- name: CreateMessage :one
INSERT INTO messages (
    id,
    channel_id,
    author_id,
    content
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING *;
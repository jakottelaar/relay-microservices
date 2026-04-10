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

-- name: GetChannelMessages :many
SELECT id, channel_id, author_id, content, created_at, updated_at
FROM messages
WHERE channel_id = sqlc.arg(channel_id)
  AND (sqlc.arg(before_id)::bigint = 0 OR id < sqlc.arg(before_id)::bigint)
  AND (sqlc.arg(after_id)::bigint = 0 OR id > sqlc.arg(after_id)::bigint)
ORDER BY id DESC
LIMIT sqlc.arg(limit_count);
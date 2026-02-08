-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts (
    id,
    email,
    password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdatePassword :exec
UPDATE accounts
SET password_hash = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: UpdateLastLogin :exec
UPDATE accounts
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;
-- name: GetAccountByEmail :one
SELECT * FROM user_accounts
WHERE email = $1 LIMIT 1;

-- name: GetAccountByID :one
SELECT * FROM user_accounts
WHERE id = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO user_accounts (
    id,
    email,
    password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdatePassword :exec
UPDATE user_accounts
SET password_hash = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: UpdateLastLogin :exec
UPDATE user_accounts
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAccount :exec
DELETE FROM user_accounts
WHERE id = $1;

-- name: CreateSession :one
INSERT INTO user_sessions (id, user_account_id, refresh_token_hash, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM user_sessions 
WHERE refresh_token_hash = $1 
  AND revoked_at IS NULL 
  AND expires_at > NOW()
LIMIT 1;

-- name: RevokeSession :exec
UPDATE user_sessions 
SET revoked_at = NOW()
WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE user_sessions 
SET revoked_at = NOW()
WHERE user_account_id = $1 AND revoked_at IS NULL;

-- name: UpdateSessionLastUsed :exec
UPDATE user_sessions 
SET last_used_at = NOW()
WHERE id = $1;

-- name: CountActiveSessions :one
SELECT COUNT(*) FROM user_sessions 
WHERE user_account_id = $1 
  AND revoked_at IS NULL 
  AND expires_at > NOW();
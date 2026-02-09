CREATE TABLE IF NOT EXISTS user_sessions (
    id BIGINT PRIMARY KEY, --Snowflake ID
    user_account_id BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,

    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT,
    ip_address INET,

    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_account_id ON user_sessions(user_account_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

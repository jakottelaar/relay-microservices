CREATE TABLE IF NOT EXISTS channels (
    id BIGINT PRIMARY KEY,
    guild_id BIGINT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    topic TEXT,
    type SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channels_guild_id ON channels(guild_id);
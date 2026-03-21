CREATE TABLE IF NOT EXISTS guilds (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    owner_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_guilds_updated_at
    BEFORE UPDATE ON guilds
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
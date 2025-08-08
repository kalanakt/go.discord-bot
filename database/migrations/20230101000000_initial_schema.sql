-- +goose Up
-- SQL in this section is executed when the migration is applied
CREATE TABLE IF NOT EXISTS command_logs (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    command_name TEXT NOT NULL,
    command_type TEXT NOT NULL,
    arguments JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_command_logs_guild_id ON command_logs(guild_id);
CREATE INDEX IF NOT EXISTS idx_command_logs_user_id ON command_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_command_logs_created_at ON command_logs(created_at);

CREATE TABLE IF NOT EXISTS interaction_events (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    interaction_type TEXT NOT NULL,
    component_id TEXT,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_interaction_events_guild_id ON interaction_events(guild_id);
CREATE INDEX IF NOT EXISTS idx_interaction_events_user_id ON interaction_events(user_id);
CREATE INDEX IF NOT EXISTS idx_interaction_events_created_at ON interaction_events(created_at);

CREATE TABLE IF NOT EXISTS bot_stats (
    id SERIAL PRIMARY KEY,
    guilds_count INTEGER NOT NULL DEFAULT 0,
    users_count INTEGER NOT NULL DEFAULT 0,
    commands_count INTEGER NOT NULL DEFAULT 0,
    uptime_seconds INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP TABLE IF EXISTS command_logs;
DROP TABLE IF EXISTS interaction_events;
DROP TABLE IF EXISTS bot_stats;
-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    config JSONB DEFAULT '{}',
    last_cursor TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_channels_bot_id FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_channels_bot_id ON channels(bot_id);
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);

-- Migrate existing iLink data from bots into channels
INSERT INTO channels (bot_id, type, status, config, last_cursor, created_at, updated_at)
SELECT
    id,
    'wechat_clawbot',
    CASE
        WHEN status = 'pending' AND qrcode_id IS NOT NULL THEN 'pending'
        WHEN status = 'active' THEN 'active'
        ELSE status
    END,
    jsonb_build_object(
        'ilink_bot_id', COALESCE(ilink_bot_id, ''),
        'ilink_user_id', COALESCE(ilink_user_id, ''),
        'bot_token', COALESCE(bot_token, ''),
        'base_url', COALESCE(base_url, ''),
        'qrcode_id', COALESCE(qrcode_id, ''),
        'qrcode_image', COALESCE(qrcode_image, '')
    ),
    last_cursor,
    created_at,
    updated_at
FROM bots;

-- Set default name for bots without one
UPDATE bots SET name = 'My Bot' WHERE name IS NULL;

-- All bots are active by default now
UPDATE bots SET status = 'active';

-- Make name NOT NULL
ALTER TABLE bots ALTER COLUMN name SET NOT NULL;

-- Drop iLink-specific columns from bots
ALTER TABLE bots DROP COLUMN IF EXISTS bot_token;
ALTER TABLE bots DROP COLUMN IF EXISTS base_url;
ALTER TABLE bots DROP COLUMN IF EXISTS ilink_bot_id;
ALTER TABLE bots DROP COLUMN IF EXISTS ilink_user_id;
ALTER TABLE bots DROP COLUMN IF EXISTS last_cursor;
ALTER TABLE bots DROP COLUMN IF EXISTS qrcode_id;
ALTER TABLE bots DROP COLUMN IF EXISTS qrcode_image;

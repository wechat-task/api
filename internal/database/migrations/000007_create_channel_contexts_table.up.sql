-- Create channel_contexts table to store user contextTokens for replying
CREATE TABLE IF NOT EXISTS channel_contexts (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    context_token TEXT NOT NULL,
    last_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_channel_contexts_channel_id FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    CONSTRAINT unique_channel_user UNIQUE (channel_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_contexts_channel_id ON channel_contexts(channel_id);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    challenge VARCHAR(255) NOT NULL,
    session_data BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    user_id INTEGER,
    username VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on challenge for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_challenge ON sessions(challenge);

-- Create index on expires_at for cleanup queries
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id) WHERE user_id IS NOT NULL;

-- Create index on username
CREATE INDEX IF NOT EXISTS idx_sessions_username ON sessions(username) WHERE username IS NOT NULL;

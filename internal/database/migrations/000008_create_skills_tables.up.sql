-- Create skills table
CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) DEFAULT '1.0.0',
    content TEXT NOT NULL,
    visibility VARCHAR(20) DEFAULT 'private',
    status VARCHAR(20) DEFAULT 'draft',
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    is_free BOOLEAN DEFAULT true,
    uses_system_llm BOOLEAN DEFAULT true,
    max_tokens INTEGER DEFAULT 2000,
    parameters JSONB DEFAULT '{}',
    schedule_cron VARCHAR(100),
    subscriber_count INTEGER DEFAULT 0,
    execution_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_skills_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create skill_subscriptions table
CREATE TABLE IF NOT EXISTS skill_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    skill_id INTEGER NOT NULL,
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    bot_id INTEGER,
    channel_id INTEGER,
    schedule_cron VARCHAR(100),
    time_zone VARCHAR(50) DEFAULT 'UTC',
    next_run_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_skill_subscriptions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_subscriptions_skill_id FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE,
    CONSTRAINT fk_skill_subscriptions_bot_id FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE SET NULL,
    CONSTRAINT fk_skill_subscriptions_channel_id FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE SET NULL,
    CONSTRAINT unique_user_skill_subscription UNIQUE (user_id, skill_id)
);

-- Create user_llm_configs table
CREATE TABLE IF NOT EXISTS user_llm_configs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    api_key TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    base_url VARCHAR(255),
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_llm_configs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create skill_execution_logs table
CREATE TABLE IF NOT EXISTS skill_execution_logs (
    id SERIAL PRIMARY KEY,
    subscription_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    input TEXT,
    output TEXT,
    error TEXT,
    token_usage INTEGER DEFAULT 0,
    duration_ms BIGINT DEFAULT 0,
    executed_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_skill_execution_logs_subscription_id FOREIGN KEY (subscription_id) REFERENCES skill_subscriptions(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_skills_user_id ON skills(user_id);
CREATE INDEX IF NOT EXISTS idx_skills_visibility_status ON skills(visibility, status);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_status ON skills(status);

CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_user_id ON skill_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_skill_id ON skill_subscriptions(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_status ON skill_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_next_run_at ON skill_subscriptions(next_run_at);
CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_bot_id ON skill_subscriptions(bot_id);
CREATE INDEX IF NOT EXISTS idx_skill_subscriptions_channel_id ON skill_subscriptions(channel_id);

CREATE INDEX IF NOT EXISTS idx_user_llm_configs_user_id ON user_llm_configs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_llm_configs_is_default ON user_llm_configs(is_default);

CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_subscription_id ON skill_execution_logs(subscription_id);
CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_executed_at ON skill_execution_logs(executed_at);
CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_status ON skill_execution_logs(status);
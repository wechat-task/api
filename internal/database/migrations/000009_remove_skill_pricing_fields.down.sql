-- Restore pricing and schedule fields to skills table (rollback)
-- This is a down migration that restores the removed columns

-- Restore is_free column with default value
ALTER TABLE skills ADD COLUMN is_free BOOLEAN DEFAULT true;

-- Restore uses_system_llm column with default value
ALTER TABLE skills ADD COLUMN uses_system_llm BOOLEAN DEFAULT true;

-- Restore max_tokens column with default value
ALTER TABLE skills ADD COLUMN max_tokens INTEGER DEFAULT 2000;

-- Restore schedule_cron column
ALTER TABLE skills ADD COLUMN schedule_cron VARCHAR(100);
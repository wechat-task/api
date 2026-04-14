-- Remove pricing and schedule fields from skills table
-- These fields have been removed from the model as pricing rules are postponed
-- and scheduling should be specified at subscription level

-- Remove is_free column (pricing postponed)
ALTER TABLE skills DROP COLUMN IF EXISTS is_free;

-- Remove uses_system_llm column (all skills use system LLM by default now)
ALTER TABLE skills DROP COLUMN IF EXISTS uses_system_llm;

-- Remove max_tokens column (token limits moved to execution config)
ALTER TABLE skills DROP COLUMN IF EXISTS max_tokens;

-- Remove schedule_cron column (scheduling should be at subscription level)
ALTER TABLE skills DROP COLUMN IF EXISTS schedule_cron;
DROP TABLE IF EXISTS demo_ai_daily_usage;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_demo;

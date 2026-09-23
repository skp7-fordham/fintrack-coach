ALTER TABLE users
    ADD COLUMN is_demo BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE demo_ai_daily_usage (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    message_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, usage_date),
    CONSTRAINT demo_ai_daily_usage_count_nonnegative CHECK (message_count >= 0)
);

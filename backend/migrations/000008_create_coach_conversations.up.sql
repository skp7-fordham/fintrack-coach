CREATE TABLE coach_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coach_conversations_user_updated
    ON coach_conversations (user_id, updated_at DESC);

CREATE TABLE coach_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL
        REFERENCES coach_conversations(id)
        ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_name TEXT,
    tool_call_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT coach_messages_role_check CHECK (
        role IN ('user', 'assistant', 'tool')
    )
);

CREATE INDEX idx_coach_messages_conversation_created
    ON coach_messages (conversation_id, created_at ASC);

CREATE INDEX idx_coach_messages_conversation_id
    ON coach_messages (conversation_id, id);

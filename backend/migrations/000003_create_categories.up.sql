CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    category_type VARCHAR(20) NOT NULL,
    color VARCHAR(20),
    icon VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT categories_category_type_check CHECK (
        category_type IN ('income', 'expense')
    ),
    CONSTRAINT categories_user_id_name_key UNIQUE (user_id, name)
);

CREATE INDEX idx_categories_user_id ON categories (user_id);

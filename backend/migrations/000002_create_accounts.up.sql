CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    account_type VARCHAR(30) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    current_balance NUMERIC(14, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT accounts_account_type_check CHECK (
        account_type IN (
            'checking',
            'savings',
            'credit_card',
            'cash',
            'investment',
            'loan'
        )
    )
);

CREATE INDEX idx_accounts_user_id ON accounts (user_id);

CREATE TABLE transaction_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    original_filename TEXT NOT NULL,
    stored_file_path TEXT NOT NULL,
    status TEXT NOT NULL,
    total_rows INTEGER NOT NULL DEFAULT 0,
    processed_rows INTEGER NOT NULL DEFAULT 0,
    successful_rows INTEGER NOT NULL DEFAULT 0,
    failed_rows INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transaction_import_jobs_status_check CHECK (
        status IN (
            'queued',
            'processing',
            'completed',
            'completed_with_errors',
            'failed'
        )
    ),
    CONSTRAINT transaction_import_jobs_counts_nonnegative CHECK (
        total_rows >= 0
        AND processed_rows >= 0
        AND successful_rows >= 0
        AND failed_rows >= 0
    )
);

CREATE INDEX idx_transaction_import_jobs_user_created
    ON transaction_import_jobs (user_id, created_at DESC);

CREATE INDEX idx_transaction_import_jobs_status
    ON transaction_import_jobs (status);

CREATE INDEX idx_transaction_import_jobs_account_id
    ON transaction_import_jobs (account_id);

CREATE TABLE transaction_import_errors (
    id BIGSERIAL PRIMARY KEY,
    import_job_id UUID NOT NULL
        REFERENCES transaction_import_jobs(id)
        ON DELETE CASCADE,
    row_number INTEGER NOT NULL,
    error_message TEXT NOT NULL,
    raw_row JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transaction_import_errors_row_number_positive CHECK (row_number > 0)
);

CREATE INDEX idx_transaction_import_errors_job_row
    ON transaction_import_errors (import_job_id, row_number);

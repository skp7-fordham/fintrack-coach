package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type ImportRepository struct {
	pool *pgxpool.Pool
}

func NewImportRepository(pool *pgxpool.Pool) *ImportRepository {
	return &ImportRepository{pool: pool}
}

func (r *ImportRepository) CreateJob(
	ctx context.Context,
	input domain.CreateTransactionImportInput,
) (*domain.TransactionImportJob, error) {
	const query = `
		INSERT INTO transaction_import_jobs (
			user_id,
			account_id,
			original_filename,
			stored_file_path,
			status
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id::text,
			user_id::text,
			account_id::text,
			original_filename,
			stored_file_path,
			status,
			total_rows,
			processed_rows,
			successful_rows,
			failed_rows,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
	`

	job, err := scanImportJob(r.pool.QueryRow(
		ctx,
		query,
		input.UserID,
		input.AccountID,
		input.OriginalFilename,
		input.StoredFilePath,
		domain.ImportStatusQueued,
	))
	if err != nil {
		return nil, fmt.Errorf("create import job: %w", err)
	}
	return job, nil
}

func (r *ImportRepository) GetJobByID(
	ctx context.Context,
	userID, jobID string,
) (*domain.TransactionImportJob, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			account_id::text,
			original_filename,
			stored_file_path,
			status,
			total_rows,
			processed_rows,
			successful_rows,
			failed_rows,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
		FROM transaction_import_jobs
		WHERE id = $1 AND user_id = $2
	`

	job, err := scanImportJob(r.pool.QueryRow(ctx, query, jobID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrImportNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get import job: %w", err)
	}
	return job, nil
}

func (r *ImportRepository) GetJobByIDInternal(
	ctx context.Context,
	jobID string,
) (*domain.TransactionImportJob, error) {
	const query = `
		SELECT
			id::text,
			user_id::text,
			account_id::text,
			original_filename,
			stored_file_path,
			status,
			total_rows,
			processed_rows,
			successful_rows,
			failed_rows,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
		FROM transaction_import_jobs
		WHERE id = $1
	`

	job, err := scanImportJob(r.pool.QueryRow(ctx, query, jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrImportNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get import job internal: %w", err)
	}
	return job, nil
}

func (r *ImportRepository) ListJobs(
	ctx context.Context,
	filter domain.ListImportJobsFilter,
) ([]domain.TransactionImportJob, int64, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM transaction_import_jobs
		WHERE user_id = $1
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, filter.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count import jobs: %w", err)
	}

	const listQuery = `
		SELECT
			id::text,
			user_id::text,
			account_id::text,
			original_filename,
			stored_file_path,
			status,
			total_rows,
			processed_rows,
			successful_rows,
			failed_rows,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
		FROM transaction_import_jobs
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, listQuery, filter.UserID, filter.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list import jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]domain.TransactionImportJob, 0)
	for rows.Next() {
		job, err := scanImportJob(rows)
		if err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate import jobs: %w", err)
	}
	return jobs, total, nil
}

func (r *ImportRepository) MarkJobProcessing(ctx context.Context, jobID string) (*domain.TransactionImportJob, error) {
	const query = `
		UPDATE transaction_import_jobs
		SET
			status = $2,
			started_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status = $3
		RETURNING
			id::text,
			user_id::text,
			account_id::text,
			original_filename,
			stored_file_path,
			status,
			total_rows,
			processed_rows,
			successful_rows,
			failed_rows,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
	`

	job, err := scanImportJob(r.pool.QueryRow(
		ctx,
		query,
		jobID,
		domain.ImportStatusProcessing,
		domain.ImportStatusQueued,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrImportAlreadyProcessing
	}
	if err != nil {
		return nil, fmt.Errorf("mark import processing: %w", err)
	}
	return job, nil
}

func (r *ImportRepository) UpdateJobProgress(
	ctx context.Context,
	jobID string,
	totalRows, processedRows, successfulRows, failedRows int,
) error {
	const query = `
		UPDATE transaction_import_jobs
		SET
			total_rows = $2,
			processed_rows = $3,
			successful_rows = $4,
			failed_rows = $5,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, jobID, totalRows, processedRows, successfulRows, failedRows)
	if err != nil {
		return fmt.Errorf("update import progress: %w", err)
	}
	return nil
}

func (r *ImportRepository) CompleteJob(
	ctx context.Context,
	jobID, status string,
	totalRows, processedRows, successfulRows, failedRows int,
	errorMessage *string,
) error {
	const query = `
		UPDATE transaction_import_jobs
		SET
			status = $2,
			total_rows = $3,
			processed_rows = $4,
			successful_rows = $5,
			failed_rows = $6,
			error_message = $7,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		jobID,
		status,
		totalRows,
		processedRows,
		successfulRows,
		failedRows,
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("complete import job: %w", err)
	}
	return nil
}

func (r *ImportRepository) FailJob(ctx context.Context, jobID, message string) error {
	const query = `
		UPDATE transaction_import_jobs
		SET
			status = $2,
			error_message = $3,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, jobID, domain.ImportStatusFailed, message)
	if err != nil {
		return fmt.Errorf("fail import job: %w", err)
	}
	return nil
}

func (r *ImportRepository) AddRowErrors(
	ctx context.Context,
	jobID string,
	errs []domain.RowValidationError,
) error {
	if len(errs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const query = `
		INSERT INTO transaction_import_errors (
			import_job_id,
			row_number,
			error_message,
			raw_row
		) VALUES ($1, $2, $3, $4::jsonb)
	`
	for _, item := range errs {
		rawJSON, err := json.Marshal(item.RawRow)
		if err != nil {
			return fmt.Errorf("marshal raw row: %w", err)
		}
		batch.Queue(query, jobID, item.RowNumber, item.ErrorMessage, rawJSON)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range errs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert import row error: %w", err)
		}
	}
	return nil
}

func (r *ImportRepository) ListErrors(
	ctx context.Context,
	filter domain.ListImportErrorsFilter,
) ([]domain.TransactionImportError, int64, error) {
	const ownershipQuery = `
		SELECT 1
		FROM transaction_import_jobs
		WHERE id = $1 AND user_id = $2
	`
	var exists int
	err := r.pool.QueryRow(ctx, ownershipQuery, filter.JobID, filter.UserID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, domain.ErrImportNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("verify import ownership: %w", err)
	}

	const countQuery = `
		SELECT COUNT(*)
		FROM transaction_import_errors
		WHERE import_job_id = $1
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, filter.JobID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count import errors: %w", err)
	}

	const listQuery = `
		SELECT
			id,
			import_job_id::text,
			row_number,
			error_message,
			raw_row,
			created_at
		FROM transaction_import_errors
		WHERE import_job_id = $1
		ORDER BY row_number ASC, id ASC
		LIMIT $2 OFFSET $3
	`
	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, listQuery, filter.JobID, filter.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list import errors: %w", err)
	}
	defer rows.Close()

	items := make([]domain.TransactionImportError, 0)
	for rows.Next() {
		var item domain.TransactionImportError
		var raw []byte
		if err := rows.Scan(
			&item.ID,
			&item.ImportJobID,
			&item.RowNumber,
			&item.ErrorMessage,
			&raw,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan import error: %w", err)
		}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &item.RawRow)
		}
		if item.RawRow == nil {
			item.RawRow = map[string]string{}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate import errors: %w", err)
	}
	return items, total, nil
}

func (r *ImportRepository) EnsureAccountOwnedByUser(ctx context.Context, userID, accountID string) error {
	const query = `
		SELECT id
		FROM accounts
		WHERE id = $1 AND user_id = $2
	`
	var id string
	err := r.pool.QueryRow(ctx, query, accountID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("verify account ownership: %w", err)
	}
	return nil
}

func (r *ImportRepository) ListUserCategories(
	ctx context.Context,
	userID string,
) ([]domain.Category, error) {
	const query = `
		SELECT
			id::text,
			name,
			category_type,
			color,
			icon,
			created_at,
			updated_at
		FROM categories
		WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CategoryType,
			&category.Color,
			&category.Icon,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (r *ImportRepository) InsertValidatedTransactions(
	ctx context.Context,
	userID, accountID string,
	rows []domain.ValidatedImportRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin import insert: %w", err)
	}
	defer tx.Rollback(ctx)

	const lockQuery = `
		SELECT id
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`
	var lockedID string
	err = tx.QueryRow(ctx, lockQuery, accountID, userID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("lock account for import: %w", err)
	}

	copySource := pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
		row := rows[i]
		return []any{
			userID,
			accountID,
			row.CategoryID,
			row.Description,
			row.Merchant,
			row.Amount,
			row.TransactionType,
			domain.TransactionStatusCompleted,
			row.TransactionDate,
			row.Notes,
		}, nil
	})

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"transactions"},
		[]string{
			"user_id",
			"account_id",
			"category_id",
			"description",
			"merchant",
			"amount",
			"transaction_type",
			"transaction_status",
			"transaction_date",
			"notes",
		},
		copySource,
	)
	if err != nil {
		return fmt.Errorf("copy import transactions: %w", err)
	}

	incomeTotal := new(big.Rat)
	expenseTotal := new(big.Rat)
	for _, row := range rows {
		amount := new(big.Rat)
		if _, ok := amount.SetString(row.Amount); !ok {
			return fmt.Errorf("invalid import amount: %s", row.Amount)
		}
		switch row.TransactionType {
		case domain.TransactionTypeIncome:
			incomeTotal.Add(incomeTotal, amount)
		case domain.TransactionTypeExpense:
			expenseTotal.Add(expenseTotal, amount)
		}
	}

	const balanceQuerySQL = `
		UPDATE accounts
		SET
			current_balance = current_balance + ($1::numeric - $2::numeric),
			updated_at = NOW()
		WHERE id = $3 AND user_id = $4
	`
	tag, err := tx.Exec(
		ctx,
		balanceQuerySQL,
		incomeTotal.FloatString(2),
		expenseTotal.FloatString(2),
		accountID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update account balance for import: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit import insert: %w", err)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanImportJob(row scannable) (*domain.TransactionImportJob, error) {
	var job domain.TransactionImportJob
	err := row.Scan(
		&job.ID,
		&job.UserID,
		&job.AccountID,
		&job.OriginalFilename,
		&job.StoredFilePath,
		&job.Status,
		&job.TotalRows,
		&job.ProcessedRows,
		&job.SuccessfulRows,
		&job.FailedRows,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

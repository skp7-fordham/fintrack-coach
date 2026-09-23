package demoseed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Seed(
	ctx context.Context,
	pool *pgxpool.Pool,
	email, password string,
	now time.Time,
) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("DEMO_USER_EMAIL is required")
	}
	if password == "" {
		return fmt.Errorf("DEMO_USER_PASSWORD is required")
	}
	if len(password) > 72 {
		return fmt.Errorf("DEMO_USER_PASSWORD must be at most 72 bytes")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo seed: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('fintrack-demo-seed'))`); err != nil {
		return fmt.Errorf("lock demo seed: %w", err)
	}

	userID, err := upsertDemoUser(ctx, tx, email, string(passwordHash))
	if err != nil {
		return err
	}

	plan := BuildPlan(email, now)
	for _, account := range plan.Accounts {
		if err := upsertAccount(ctx, tx, userID, account); err != nil {
			return err
		}
	}
	for _, category := range plan.Categories {
		if err := upsertCategory(ctx, tx, userID, category); err != nil {
			return err
		}
	}
	for _, transaction := range plan.Transactions {
		if err := upsertTransaction(ctx, tx, userID, transaction); err != nil {
			return err
		}
	}
	for _, importJob := range plan.Imports {
		if err := upsertImport(ctx, tx, userID, importJob); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo seed: %w", err)
	}
	return nil
}

func upsertDemoUser(ctx context.Context, tx pgx.Tx, email, passwordHash string) (string, error) {
	const selectQuery = `
		SELECT id::text
		FROM users
		WHERE LOWER(email) = $1
		FOR UPDATE
	`
	var userID string
	err := tx.QueryRow(ctx, selectQuery, email).Scan(&userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		const insertQuery = `
			INSERT INTO users (email, password_hash, is_demo)
			VALUES ($1, $2, TRUE)
			RETURNING id::text
		`
		if err := tx.QueryRow(ctx, insertQuery, email, passwordHash).Scan(&userID); err != nil {
			return "", fmt.Errorf("create demo user: %w", err)
		}
	case err != nil:
		return "", fmt.Errorf("find demo user: %w", err)
	default:
		const updateQuery = `
			UPDATE users
			SET email = $2,
			    password_hash = $3,
			    is_demo = TRUE,
			    updated_at = NOW()
			WHERE id = $1
		`
		if _, err := tx.Exec(ctx, updateQuery, userID, email, passwordHash); err != nil {
			return "", fmt.Errorf("update demo user: %w", err)
		}
	}
	return userID, nil
}

func upsertAccount(ctx context.Context, tx pgx.Tx, userID string, account AccountSeed) error {
	const query = `
		INSERT INTO accounts (
			id, user_id, name, account_type, currency, current_balance
		) VALUES ($1, $2, $3, $4, $5, $6::numeric)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			name = EXCLUDED.name,
			account_type = EXCLUDED.account_type,
			currency = EXCLUDED.currency,
			current_balance = EXCLUDED.current_balance,
			updated_at = NOW()
	`
	if _, err := tx.Exec(
		ctx,
		query,
		account.ID,
		userID,
		account.Name,
		account.AccountType,
		account.Currency,
		account.CurrentBalance,
	); err != nil {
		return fmt.Errorf("upsert demo account %q: %w", account.Name, err)
	}
	return nil
}

func upsertCategory(ctx context.Context, tx pgx.Tx, userID string, category CategorySeed) error {
	const query = `
		INSERT INTO categories (
			id, user_id, name, category_type, color, icon
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			name = EXCLUDED.name,
			category_type = EXCLUDED.category_type,
			color = EXCLUDED.color,
			icon = EXCLUDED.icon,
			updated_at = NOW()
	`
	if _, err := tx.Exec(
		ctx,
		query,
		category.ID,
		userID,
		category.Name,
		category.CategoryType,
		category.Color,
		category.Icon,
	); err != nil {
		return fmt.Errorf("upsert demo category %q: %w", category.Name, err)
	}
	return nil
}

func upsertTransaction(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	transaction TransactionSeed,
) error {
	const query = `
		INSERT INTO transactions (
			id, user_id, account_id, category_id, description, merchant,
			amount, transaction_type, transaction_status, transaction_date, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::numeric, $8, $9, $10, $11
		)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			account_id = EXCLUDED.account_id,
			category_id = EXCLUDED.category_id,
			description = EXCLUDED.description,
			merchant = EXCLUDED.merchant,
			amount = EXCLUDED.amount,
			transaction_type = EXCLUDED.transaction_type,
			transaction_status = EXCLUDED.transaction_status,
			transaction_date = EXCLUDED.transaction_date,
			notes = EXCLUDED.notes,
			updated_at = NOW()
	`
	if _, err := tx.Exec(
		ctx,
		query,
		transaction.ID,
		userID,
		transaction.AccountID,
		transaction.CategoryID,
		transaction.Description,
		transaction.Merchant,
		transaction.Amount,
		transaction.TransactionType,
		transaction.TransactionStatus,
		transaction.TransactionDate,
		transaction.Notes,
	); err != nil {
		return fmt.Errorf("upsert demo transaction %q: %w", transaction.Description, err)
	}
	return nil
}

func upsertImport(ctx context.Context, tx pgx.Tx, userID string, job ImportSeed) error {
	const query = `
		INSERT INTO transaction_import_jobs (
			id, user_id, account_id, original_filename, stored_file_path, status,
			total_rows, processed_rows, successful_rows, failed_rows,
			started_at, completed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $7, $8, $9,
			$10, $11, $12, $12
		)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			account_id = EXCLUDED.account_id,
			original_filename = EXCLUDED.original_filename,
			stored_file_path = EXCLUDED.stored_file_path,
			status = EXCLUDED.status,
			total_rows = EXCLUDED.total_rows,
			processed_rows = EXCLUDED.processed_rows,
			successful_rows = EXCLUDED.successful_rows,
			failed_rows = EXCLUDED.failed_rows,
			error_message = NULL,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
	`
	startedAt := job.CreatedAt.Add(time.Minute)
	completedAt := job.CreatedAt.Add(2 * time.Minute)
	if _, err := tx.Exec(
		ctx,
		query,
		job.ID,
		userID,
		job.AccountID,
		job.OriginalFilename,
		"demo-history://"+job.OriginalFilename,
		job.Status,
		job.TotalRows,
		job.SuccessfulRows,
		job.FailedRows,
		startedAt,
		completedAt,
		job.CreatedAt,
	); err != nil {
		return fmt.Errorf("upsert demo import %q: %w", job.OriginalFilename, err)
	}

	if _, err := tx.Exec(
		ctx,
		`DELETE FROM transaction_import_errors WHERE import_job_id = $1`,
		job.ID,
	); err != nil {
		return fmt.Errorf("clear demo import errors %q: %w", job.OriginalFilename, err)
	}

	const errorQuery = `
		INSERT INTO transaction_import_errors (
			import_job_id, row_number, error_message, raw_row
		) VALUES ($1, $2, $3, $4::jsonb)
	`
	for _, rowError := range job.Errors {
		rawJSON, err := json.Marshal(rowError.RawRow)
		if err != nil {
			return fmt.Errorf("encode demo import error: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			errorQuery,
			job.ID,
			rowError.RowNumber,
			rowError.ErrorMessage,
			string(rawJSON),
		); err != nil {
			return fmt.Errorf("insert demo import error: %w", err)
		}
	}
	return nil
}

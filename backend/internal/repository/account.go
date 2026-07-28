package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) CreateAccount(
	ctx context.Context,
	input domain.CreateAccountInput,
) (*domain.Account, error) {
	const query = `
		INSERT INTO accounts (user_id, name, account_type, currency, current_balance)
		VALUES ($1, $2, $3, $4, $5::numeric)
		RETURNING
			id::text,
			name,
			account_type,
			currency,
			current_balance::numeric(14, 2)::text,
			created_at,
			updated_at
	`

	var account domain.Account
	err := r.pool.QueryRow(
		ctx,
		query,
		input.UserID,
		input.Name,
		input.AccountType,
		input.Currency,
		input.InitialBalance,
	).Scan(
		&account.ID,
		&account.Name,
		&account.AccountType,
		&account.Currency,
		&account.CurrentBalance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) ListAccounts(ctx context.Context, userID string) ([]domain.Account, error) {
	const query = `
		SELECT
			id::text,
			name,
			account_type,
			currency,
			current_balance::numeric(14, 2)::text,
			created_at,
			updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0)
	for rows.Next() {
		var account domain.Account
		if err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.AccountType,
			&account.Currency,
			&account.CurrentBalance,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

func (r *AccountRepository) UpdateAccount(
	ctx context.Context,
	input domain.UpdateAccountInput,
) (*domain.Account, error) {
	const query = `
		UPDATE accounts
		SET
			name = COALESCE($3, name),
			account_type = COALESCE($4, account_type),
			currency = COALESCE($5, currency),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING
			id::text,
			name,
			account_type,
			currency,
			current_balance::numeric(14, 2)::text,
			created_at,
			updated_at
	`

	var account domain.Account
	err := r.pool.QueryRow(
		ctx,
		query,
		input.AccountID,
		input.UserID,
		input.Name,
		input.AccountType,
		input.Currency,
	).Scan(
		&account.ID,
		&account.Name,
		&account.AccountType,
		&account.Currency,
		&account.CurrentBalance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) DeleteAccount(ctx context.Context, userID, accountID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete account: %w", err)
	}
	defer tx.Rollback(ctx)

	const lockQuery = `
		SELECT id
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`
	var id string
	err = tx.QueryRow(ctx, lockQuery, accountID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}

	const countQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM transactions
			WHERE account_id = $1
		)
	`
	var hasTransactions bool
	if err := tx.QueryRow(ctx, countQuery, accountID).Scan(&hasTransactions); err != nil {
		return fmt.Errorf("check account transactions: %w", err)
	}
	if hasTransactions {
		return domain.ErrAccountHasTransactions
	}

	const deleteQuery = `
		DELETE FROM accounts
		WHERE id = $1 AND user_id = $2
	`
	tag, err := tx.Exec(ctx, deleteQuery, accountID, userID)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete account: %w", err)
	}
	return nil
}

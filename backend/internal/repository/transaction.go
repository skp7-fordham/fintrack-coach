package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) CreateTransaction(
	ctx context.Context,
	input domain.CreateTransactionInput,
) (*domain.Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after commit

	if err := r.ensureAccountOwnedByUser(ctx, tx, input.AccountID, input.UserID); err != nil {
		return nil, err
	}

	if input.CategoryID != nil {
		if err := r.ensureCategoryOwnedByUser(ctx, tx, *input.CategoryID, input.UserID); err != nil {
			return nil, err
		}
	}

	txn, err := r.insertTransaction(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := r.applyBalanceUpdate(ctx, tx, input); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return txn, nil
}

func (r *TransactionRepository) ensureAccountOwnedByUser(
	ctx context.Context,
	tx pgx.Tx,
	accountID, userID string,
) error {
	const query = `
		SELECT id
		FROM accounts
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`

	var id string
	err := tx.QueryRow(ctx, query, accountID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("verify account ownership: %w", err)
	}
	return nil
}

func (r *TransactionRepository) ensureCategoryOwnedByUser(
	ctx context.Context,
	tx pgx.Tx,
	categoryID, userID string,
) error {
	const query = `
		SELECT id
		FROM categories
		WHERE id = $1 AND user_id = $2
	`

	var id string
	err := tx.QueryRow(ctx, query, categoryID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrCategoryNotFound
	}
	if err != nil {
		return fmt.Errorf("verify category ownership: %w", err)
	}
	return nil
}

func (r *TransactionRepository) insertTransaction(
	ctx context.Context,
	tx pgx.Tx,
	input domain.CreateTransactionInput,
) (*domain.Transaction, error) {
	const query = `
		INSERT INTO transactions (
			user_id,
			account_id,
			category_id,
			description,
			merchant,
			amount,
			transaction_type,
			transaction_status,
			transaction_date,
			notes
		) VALUES (
			$1, $2, $3, $4, $5, $6::numeric, $7, $8, $9, $10
		)
		RETURNING
			id::text,
			user_id::text,
			account_id::text,
			category_id::text,
			description,
			merchant,
			amount::text,
			transaction_type,
			transaction_status,
			transaction_date,
			notes,
			created_at,
			updated_at
	`

	var txn domain.Transaction
	err := tx.QueryRow(
		ctx,
		query,
		input.UserID,
		input.AccountID,
		input.CategoryID,
		input.Description,
		input.Merchant,
		input.Amount,
		input.TransactionType,
		input.TransactionStatus,
		input.TransactionDate,
		input.Notes,
	).Scan(
		&txn.ID,
		&txn.UserID,
		&txn.AccountID,
		&txn.CategoryID,
		&txn.Description,
		&txn.Merchant,
		&txn.Amount,
		&txn.TransactionType,
		&txn.TransactionStatus,
		&txn.TransactionDate,
		&txn.Notes,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	return &txn, nil
}

func (r *TransactionRepository) applyBalanceUpdate(
	ctx context.Context,
	tx pgx.Tx,
	input domain.CreateTransactionInput,
) error {
	if input.TransactionStatus != domain.TransactionStatusCompleted {
		return nil
	}

	var delta string
	switch input.TransactionType {
	case domain.TransactionTypeIncome:
		delta = input.Amount
	case domain.TransactionTypeExpense:
		delta = "-" + input.Amount
	case domain.TransactionTypeTransfer:
		return nil
	default:
		return fmt.Errorf("unsupported transaction type for balance update: %s", input.TransactionType)
	}

	const query = `
		UPDATE accounts
		SET current_balance = current_balance + $1::numeric,
		    updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`

	tag, err := tx.Exec(ctx, query, delta, input.AccountID, input.UserID)
	if err != nil {
		return fmt.Errorf("update account balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

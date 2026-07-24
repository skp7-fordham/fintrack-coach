package domain

import (
	"errors"
	"time"
)

const (
	TransactionTypeIncome   = "income"
	TransactionTypeExpense  = "expense"
	TransactionTypeTransfer = "transfer"

	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

var (
	ErrAccountNotFound  = errors.New("account not found")
	ErrCategoryNotFound = errors.New("category not found")
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Transaction struct {
	ID                string
	UserID            string
	AccountID         string
	CategoryID        *string
	Description       string
	Merchant          *string
	Amount            string
	TransactionType   string
	TransactionStatus string
	TransactionDate   time.Time
	Notes             *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreateTransactionInput struct {
	UserID            string
	AccountID         string
	CategoryID        *string
	Description       string
	Merchant          *string
	Amount            string
	TransactionType   string
	TransactionStatus string
	TransactionDate   time.Time
	Notes             *string
}

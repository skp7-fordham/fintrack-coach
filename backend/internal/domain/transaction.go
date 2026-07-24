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

const (
	SortByTransactionDate = "transaction_date"
	SortByCreatedAt       = "created_at"

	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// TransactionFilter holds validated list filters for GET /transactions.
type TransactionFilter struct {
	UserID            string
	AccountID         *string
	CategoryID        *string
	TransactionType   *string
	TransactionStatus *string
	From              *time.Time
	To                *time.Time
	Search            string
	Page              int
	PageSize          int
	Sort              string
	Order             string
}

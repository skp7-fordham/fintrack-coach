package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

// amountPattern matches positive decimal amounts that fit NUMERIC(14,2):
// up to 12 digits before the decimal point and up to 2 after.
var amountPattern = regexp.MustCompile(`^[0-9]{1,12}(\.[0-9]{1,2})?$`)

type transactionRepository interface {
	CreateTransaction(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error)
}

type TransactionService struct {
	repo transactionRepository
}

func NewTransactionService(repo transactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(
	ctx context.Context,
	input domain.CreateTransactionInput,
) (*domain.Transaction, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.AccountID = strings.TrimSpace(input.AccountID)
	input.Description = strings.TrimSpace(input.Description)
	input.Amount = strings.TrimSpace(input.Amount)

	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	txn, err := s.repo.CreateTransaction(ctx, input)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

func validateCreateInput(input domain.CreateTransactionInput) error {
	if !isValidUUID(input.UserID) {
		return &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(input.AccountID) {
		return &domain.ValidationError{Message: "account_id must be a valid UUID"}
	}
	if input.CategoryID != nil {
		if !isValidUUID(*input.CategoryID) {
			return &domain.ValidationError{Message: "category_id must be a valid UUID"}
		}
	}

	if input.Description == "" {
		return &domain.ValidationError{Message: "description is required"}
	}

	if err := validateAmount(input.Amount); err != nil {
		return err
	}

	switch input.TransactionType {
	case domain.TransactionTypeIncome, domain.TransactionTypeExpense, domain.TransactionTypeTransfer:
	default:
		return &domain.ValidationError{Message: "transaction_type must be income, expense, or transfer"}
	}

	switch input.TransactionStatus {
	case domain.TransactionStatusPending, domain.TransactionStatusCompleted, domain.TransactionStatusFailed:
	default:
		return &domain.ValidationError{Message: "transaction_status must be pending, completed, or failed"}
	}

	if input.TransactionDate.IsZero() {
		return &domain.ValidationError{Message: "transaction_date must use YYYY-MM-DD"}
	}

	return nil
}

func validateAmount(amount string) error {
	if amount == "" {
		return &domain.ValidationError{Message: "amount must be greater than zero"}
	}

	// Reject scientific notation, signs, fractions, and other non-decimal forms.
	if !amountPattern.MatchString(amount) {
		return &domain.ValidationError{
			Message: "amount must be a positive decimal with at most 12 digits before the decimal and 2 after",
		}
	}

	if isAllZeroDecimal(amount) {
		return &domain.ValidationError{Message: "amount must be greater than zero"}
	}

	return nil
}

func isAllZeroDecimal(amount string) bool {
	for _, c := range amount {
		if c >= '1' && c <= '9' {
			return false
		}
	}
	return true
}

func isValidUUID(value string) bool {
	if len(value) != 36 {
		return false
	}

	// 8-4-4-4-12 hex form
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHex(c) {
				return false
			}
		}
	}
	return true
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

// ParseTransactionDate parses a YYYY-MM-DD date into UTC midnight.
func ParseTransactionDate(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("transaction_date must use YYYY-MM-DD")
	}
	return t, nil
}

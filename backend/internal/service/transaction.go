package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
	maxSearchLength = 100
)

// amountPattern matches positive decimal amounts that fit NUMERIC(14,2):
// up to 12 digits before the decimal point and up to 2 after.
var amountPattern = regexp.MustCompile(`^[0-9]{1,12}(\.[0-9]{1,2})?$`)

type transactionRepository interface {
	CreateTransaction(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error)
	ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, int64, error)
}

type TransactionService struct {
	repo transactionRepository
}

func NewTransactionService(repo transactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

type ListTransactionsResult struct {
	Transactions []domain.Transaction
	Page         int
	PageSize     int
	TotalItems   int64
	TotalPages   int
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

func (s *TransactionService) ListTransactions(
	ctx context.Context,
	query dto.ListTransactionsQuery,
) (*ListTransactionsResult, error) {
	filter, err := buildTransactionFilter(query)
	if err != nil {
		return nil, err
	}

	transactions, totalItems, err := s.repo.ListTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int((totalItems + int64(filter.PageSize) - 1) / int64(filter.PageSize))
	}

	return &ListTransactionsResult{
		Transactions: transactions,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
	}, nil
}

func buildTransactionFilter(query dto.ListTransactionsQuery) (domain.TransactionFilter, error) {
	userID := strings.TrimSpace(query.UserID)
	if !isValidUUID(userID) {
		return domain.TransactionFilter{}, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	filter := domain.TransactionFilter{
		UserID:   userID,
		Page:     defaultPage,
		PageSize: defaultPageSize,
		Sort:     domain.SortByTransactionDate,
		Order:    domain.OrderDesc,
	}

	if accountID := strings.TrimSpace(query.AccountID); accountID != "" {
		if !isValidUUID(accountID) {
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "account_id must be a valid UUID"}
		}
		filter.AccountID = &accountID
	}

	if categoryID := strings.TrimSpace(query.CategoryID); categoryID != "" {
		if !isValidUUID(categoryID) {
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "category_id must be a valid UUID"}
		}
		filter.CategoryID = &categoryID
	}

	if txType := strings.TrimSpace(query.TransactionType); txType != "" {
		switch txType {
		case domain.TransactionTypeIncome, domain.TransactionTypeExpense, domain.TransactionTypeTransfer:
			filter.TransactionType = &txType
		default:
			return domain.TransactionFilter{}, &domain.ValidationError{
				Message: "transaction_type must be income, expense, or transfer",
			}
		}
	}

	if status := strings.TrimSpace(query.TransactionStatus); status != "" {
		switch status {
		case domain.TransactionStatusPending, domain.TransactionStatusCompleted, domain.TransactionStatusFailed:
			filter.TransactionStatus = &status
		default:
			return domain.TransactionFilter{}, &domain.ValidationError{
				Message: "transaction_status must be pending, completed, or failed",
			}
		}
	}

	fromRaw := strings.TrimSpace(query.From)
	toRaw := strings.TrimSpace(query.To)
	var fromDate, toDate time.Time
	var hasFrom, hasTo bool

	if fromRaw != "" {
		parsed, err := ParseTransactionDate(fromRaw)
		if err != nil {
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "from must use YYYY-MM-DD"}
		}
		fromDate = parsed
		hasFrom = true
		filter.From = &fromDate
	}

	if toRaw != "" {
		parsed, err := ParseTransactionDate(toRaw)
		if err != nil {
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "to must use YYYY-MM-DD"}
		}
		toDate = parsed
		hasTo = true
		filter.To = &toDate
	}

	if hasFrom && hasTo && fromDate.After(toDate) {
		return domain.TransactionFilter{}, &domain.ValidationError{Message: "from must not be later than to"}
	}

	search := strings.TrimSpace(query.Search)
	if utf8.RuneCountInString(search) > maxSearchLength {
		return domain.TransactionFilter{}, &domain.ValidationError{
			Message: fmt.Sprintf("search must be at most %d characters", maxSearchLength),
		}
	}
	filter.Search = search

	if pageRaw := strings.TrimSpace(query.Page); pageRaw != "" {
		page, err := strconv.Atoi(pageRaw)
		if err != nil || page < 1 {
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "page must be a positive integer"}
		}
		filter.Page = page
	}

	if pageSizeRaw := strings.TrimSpace(query.PageSize); pageSizeRaw != "" {
		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil || pageSize < 1 || pageSize > maxPageSize {
			return domain.TransactionFilter{}, &domain.ValidationError{
				Message: fmt.Sprintf("page_size must be an integer between 1 and %d", maxPageSize),
			}
		}
		filter.PageSize = pageSize
	}

	if sort := strings.TrimSpace(query.Sort); sort != "" {
		switch sort {
		case domain.SortByTransactionDate, domain.SortByCreatedAt:
			filter.Sort = sort
		default:
			return domain.TransactionFilter{}, &domain.ValidationError{
				Message: "sort must be transaction_date or created_at",
			}
		}
	}

	if order := strings.ToLower(strings.TrimSpace(query.Order)); order != "" {
		switch order {
		case domain.OrderAsc, domain.OrderDesc:
			filter.Order = order
		default:
			return domain.TransactionFilter{}, &domain.ValidationError{Message: "order must be asc or desc"}
		}
	}

	return filter, nil
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

package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
)

const (
	maxAccountNameLength  = 100
	maxCategoryNameLength = 100
	maxCategoryIconLength = 50
)

var (
	signedAmountPattern = regexp.MustCompile(`^-?[0-9]{1,12}(\.[0-9]{1,2})?$`)
	currencyPattern     = regexp.MustCompile(`^[A-Za-z]{3}$`)
	hexColorPattern     = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

var allowedAccountTypes = map[string]struct{}{
	"checking":    {},
	"savings":     {},
	"credit_card": {},
	"cash":        {},
	"investment":  {},
	"loan":        {},
}

type accountRepository interface {
	CreateAccount(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error)
	ListAccounts(ctx context.Context, userID string) ([]domain.Account, error)
	UpdateAccount(ctx context.Context, input domain.UpdateAccountInput) (*domain.Account, error)
	DeleteAccount(ctx context.Context, userID, accountID string) error
}

type AccountService struct {
	repo accountRepository
}

func NewAccountService(repo accountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(
	ctx context.Context,
	userID string,
	req dto.CreateAccountRequest,
) (*domain.Account, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	name, err := validateAccountName(req.Name)
	if err != nil {
		return nil, err
	}

	accountType, err := validateAccountType(req.AccountType)
	if err != nil {
		return nil, err
	}

	currency, err := validateCurrency(req.Currency)
	if err != nil {
		return nil, err
	}

	initialBalance, err := parseOptionalSignedAmount(req.InitialBalance, "0.00")
	if err != nil {
		return nil, err
	}

	return s.repo.CreateAccount(ctx, domain.CreateAccountInput{
		UserID:         userID,
		Name:           name,
		AccountType:    accountType,
		Currency:       currency,
		InitialBalance: initialBalance,
	})
}

func (s *AccountService) ListAccounts(ctx context.Context, userID string) ([]domain.Account, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	return s.repo.ListAccounts(ctx, userID)
}

func (s *AccountService) UpdateAccount(
	ctx context.Context,
	userID, accountID string,
	req dto.UpdateAccountRequest,
) (*domain.Account, error) {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(accountID) {
		return nil, &domain.ValidationError{Message: "account id must be a valid UUID"}
	}

	if req.Name == nil && req.AccountType == nil && req.Currency == nil {
		return nil, &domain.ValidationError{Message: "at least one field is required"}
	}

	input := domain.UpdateAccountInput{
		UserID:    userID,
		AccountID: accountID,
	}

	if req.Name != nil {
		name, err := validateAccountName(*req.Name)
		if err != nil {
			return nil, err
		}
		input.Name = &name
	}
	if req.AccountType != nil {
		accountType, err := validateAccountType(*req.AccountType)
		if err != nil {
			return nil, err
		}
		input.AccountType = &accountType
	}
	if req.Currency != nil {
		currency, err := validateCurrency(*req.Currency)
		if err != nil {
			return nil, err
		}
		input.Currency = &currency
	}

	return s.repo.UpdateAccount(ctx, input)
}

func (s *AccountService) DeleteAccount(ctx context.Context, userID, accountID string) error {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)
	if !isValidUUID(userID) {
		return &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(accountID) {
		return &domain.ValidationError{Message: "account id must be a valid UUID"}
	}
	return s.repo.DeleteAccount(ctx, userID, accountID)
}

func validateAccountName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", &domain.ValidationError{Message: "name is required"}
	}
	if utf8.RuneCountInString(name) > maxAccountNameLength {
		return "", &domain.ValidationError{Message: "name must be at most 100 characters"}
	}
	return name, nil
}

func validateAccountType(raw string) (string, error) {
	accountType := strings.TrimSpace(raw)
	if _, ok := allowedAccountTypes[accountType]; !ok {
		return "", &domain.ValidationError{
			Message: "account_type must be checking, savings, credit_card, cash, investment, or loan",
		}
	}
	return accountType, nil
}

func validateCurrency(raw string) (string, error) {
	currency := strings.TrimSpace(raw)
	if !currencyPattern.MatchString(currency) {
		return "", &domain.ValidationError{Message: "currency must be exactly 3 alphabetic characters"}
	}
	return strings.ToUpper(currency), nil
}

func parseOptionalSignedAmount(value *json.Number, defaultValue string) (string, error) {
	if value == nil {
		return defaultValue, nil
	}
	amount := strings.TrimSpace(value.String())
	if amount == "" {
		return "", &domain.ValidationError{Message: "initial_balance must be a valid decimal"}
	}
	if !signedAmountPattern.MatchString(amount) {
		return "", &domain.ValidationError{
			Message: "initial_balance must be a decimal with at most 12 digits before the decimal and 2 after",
		}
	}
	return amount, nil
}

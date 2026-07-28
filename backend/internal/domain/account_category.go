package domain

import (
	"errors"
	"time"
)

var (
	ErrAccountHasTransactions = errors.New("account has transactions and cannot be deleted")
	ErrCategoryAlreadyExists  = errors.New("category already exists")
)

// Account is the public account representation.
type Account struct {
	ID             string
	Name           string
	AccountType    string
	Currency       string
	CurrentBalance string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateAccountInput struct {
	UserID         string
	Name           string
	AccountType    string
	Currency       string
	InitialBalance string
}

type UpdateAccountInput struct {
	UserID      string
	AccountID   string
	Name        *string
	AccountType *string
	Currency    *string
}

// Category is the public category representation.
type Category struct {
	ID           string
	Name         string
	CategoryType string
	Color        *string
	Icon         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateCategoryInput struct {
	UserID       string
	Name         string
	CategoryType string
	Color        *string
	Icon         *string
}

type UpdateCategoryInput struct {
	UserID       string
	CategoryID   string
	Name         *string
	CategoryType *string
	Color        *string
	Icon         *string
}

type ListCategoriesFilter struct {
	UserID       string
	CategoryType *string
}

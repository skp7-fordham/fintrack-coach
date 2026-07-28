package dto

import "encoding/json"

// CreateAccountRequest is the HTTP body for POST /accounts.
type CreateAccountRequest struct {
	Name           string       `json:"name"`
	AccountType    string       `json:"account_type"`
	Currency       string       `json:"currency"`
	InitialBalance *json.Number `json:"initial_balance"`
}

// UpdateAccountRequest is the HTTP body for PATCH /accounts/{id}.
type UpdateAccountRequest struct {
	Name        *string `json:"name"`
	AccountType *string `json:"account_type"`
	Currency    *string `json:"currency"`
}

// CreateCategoryRequest is the HTTP body for POST /categories.
type CreateCategoryRequest struct {
	Name         string  `json:"name"`
	CategoryType string  `json:"category_type"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
}

// UpdateCategoryRequest is the HTTP body for PATCH /categories/{id}.
type UpdateCategoryRequest struct {
	Name         *string `json:"name"`
	CategoryType *string `json:"category_type"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
}

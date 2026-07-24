package dto

import "encoding/json"

// CreateTransactionRequest is the HTTP body for POST /transactions.
type CreateTransactionRequest struct {
	UserID            string          `json:"user_id"`
	AccountID         string          `json:"account_id"`
	CategoryID        *string         `json:"category_id"`
	Description       string          `json:"description"`
	Merchant          *string         `json:"merchant"`
	Amount            json.Number     `json:"amount"`
	TransactionType   string          `json:"transaction_type"`
	TransactionStatus string          `json:"transaction_status"`
	TransactionDate   string          `json:"transaction_date"`
	Notes             *string         `json:"notes"`
}

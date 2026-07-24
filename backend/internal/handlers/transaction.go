package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

const maxTransactionBodyBytes = 1 << 20 // 1 MiB

type TransactionHandler struct {
	service *service.TransactionService
	logger  *slog.Logger
}

func NewTransactionHandler(svc *service.TransactionService, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		service: svc,
		logger:  logger,
	}
}

type createTransactionResponse struct {
	Data transactionResponseData `json:"data"`
}

type transactionResponseData struct {
	ID                string  `json:"id"`
	UserID            string  `json:"user_id"`
	AccountID         string  `json:"account_id"`
	CategoryID        *string `json:"category_id"`
	Description       string  `json:"description"`
	Merchant          *string `json:"merchant"`
	Amount            string  `json:"amount"`
	TransactionType   string  `json:"transaction_type"`
	TransactionStatus string  `json:"transaction_status"`
	TransactionDate   string  `json:"transaction_date"`
	Notes             *string `json:"notes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxTransactionBodyBytes)

	var req dto.CreateTransactionRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	amountStr := strings.TrimSpace(req.Amount.String())
	if amountStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "amount must be a valid number"})
		return
	}

	txDate, err := service.ParseTransactionDate(req.TransactionDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	input := domain.CreateTransactionInput{
		UserID:            req.UserID,
		AccountID:         req.AccountID,
		CategoryID:        req.CategoryID,
		Description:       req.Description,
		Merchant:          req.Merchant,
		Amount:            amountStr,
		TransactionType:   req.TransactionType,
		TransactionStatus: req.TransactionStatus,
		TransactionDate:   txDate,
		Notes:             req.Notes,
	}

	txn, err := h.service.CreateTransaction(r.Context(), input)
	if err != nil {
		h.writeCreateError(w, err)
		return
	}

	h.logger.Info(
		"transaction created",
		"transaction_id", txn.ID,
		"user_id", txn.UserID,
		"account_id", txn.AccountID,
		"transaction_type", txn.TransactionType,
		"transaction_status", txn.TransactionStatus,
	)

	writeJSON(w, http.StatusCreated, createTransactionResponse{
		Data: toTransactionResponseData(txn),
	})
}

func (h *TransactionHandler) writeCreateError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrAccountNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "account not found for user"})
	case errors.Is(err, domain.ErrCategoryNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "category not found for user"})
	default:
		h.logger.Error("failed to create transaction", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toTransactionResponseData(txn *domain.Transaction) transactionResponseData {
	return transactionResponseData{
		ID:                txn.ID,
		UserID:            txn.UserID,
		AccountID:         txn.AccountID,
		CategoryID:        txn.CategoryID,
		Description:       txn.Description,
		Merchant:          txn.Merchant,
		Amount:            txn.Amount,
		TransactionType:   txn.TransactionType,
		TransactionStatus: txn.TransactionStatus,
		TransactionDate:   txn.TransactionDate.Format("2006-01-02"),
		Notes:             txn.Notes,
		CreatedAt:         txn.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:         txn.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func decodeJSONStrict(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return errors.New("request body too large")
		}
		return errors.New("invalid JSON body")
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

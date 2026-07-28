package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

const maxAccountBodyBytes = 1 << 20 // 1 MiB

type AccountHandler struct {
	service *service.AccountService
	logger  *slog.Logger
}

func NewAccountHandler(svc *service.AccountService, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{service: svc, logger: logger}
}

type accountResponse struct {
	Data accountData `json:"data"`
}

type accountListResponse struct {
	Data []accountData `json:"data"`
}

type accountData struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AccountType    string `json:"account_type"`
	Currency       string `json:"currency"`
	CurrentBalance string `json:"current_balance"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAccountBodyBytes)

	var req dto.CreateAccountRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	account, err := h.service.CreateAccount(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err, "failed to create account")
		return
	}

	h.logger.Info("account created", "user_id", userID, "account_id", account.ID)
	writeJSON(w, http.StatusCreated, accountResponse{Data: toAccountData(account)})
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	accounts, err := h.service.ListAccounts(r.Context(), userID)
	if err != nil {
		h.writeError(w, err, "failed to list accounts")
		return
	}

	h.logger.Info("accounts listed", "user_id", userID, "count", len(accounts))

	data := make([]accountData, 0, len(accounts))
	for i := range accounts {
		data = append(data, toAccountData(&accounts[i]))
	}
	writeJSON(w, http.StatusOK, accountListResponse{Data: data})
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	accountID := r.PathValue("id")
	if !isPathUUID(accountID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "account id must be a valid UUID"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAccountBodyBytes)

	var req dto.UpdateAccountRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	account, err := h.service.UpdateAccount(r.Context(), userID, accountID, req)
	if err != nil {
		h.writeError(w, err, "failed to update account")
		return
	}

	h.logger.Info("account updated", "user_id", userID, "account_id", account.ID)
	writeJSON(w, http.StatusOK, accountResponse{Data: toAccountData(account)})
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	accountID := r.PathValue("id")
	if !isPathUUID(accountID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "account id must be a valid UUID"})
		return
	}

	if err := h.service.DeleteAccount(r.Context(), userID, accountID); err != nil {
		h.writeError(w, err, "failed to delete account")
		return
	}

	h.logger.Info("account deleted", "user_id", userID, "account_id", accountID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) writeError(w http.ResponseWriter, err error, logMessage string) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrAccountNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "account not found"})
	case errors.Is(err, domain.ErrAccountHasTransactions):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "account has transactions and cannot be deleted"})
	default:
		h.logger.Error(logMessage, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toAccountData(account *domain.Account) accountData {
	return accountData{
		ID:             account.ID,
		Name:           account.Name,
		AccountType:    account.AccountType,
		Currency:       account.Currency,
		CurrentBalance: account.CurrentBalance,
		CreatedAt:      account.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:      account.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func isPathUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

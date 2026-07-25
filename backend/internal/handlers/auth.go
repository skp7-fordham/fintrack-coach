package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/dto"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

const maxAuthBodyBytes = 1 << 20 // 1 MiB

type AuthHandler struct {
	service *service.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(svc *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		service: svc,
		logger:  logger,
	}
}

type authSuccessResponse struct {
	Data authSuccessData `json:"data"`
}

type authSuccessData struct {
	User        authUserData `json:"user"`
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
}

type authUserData struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req dto.RegisterRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.writeRegisterError(w, err)
		return
	}

	h.logger.Info("user registered", "user_id", result.User.ID)
	writeJSON(w, http.StatusCreated, toAuthSuccessResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req dto.LoginRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.writeLoginError(w, err)
		return
	}

	h.logger.Info("user logged in", "user_id", result.User.ID)
	writeJSON(w, http.StatusOK, toAuthSuccessResponse(result))
}

func (h *AuthHandler) writeRegisterError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "email already exists"})
	default:
		h.logger.Error("registration failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func (h *AuthHandler) writeLoginError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
	default:
		h.logger.Error("login failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toAuthSuccessResponse(result *domain.AuthResult) authSuccessResponse {
	return authSuccessResponse{
		Data: authSuccessData{
			User: authUserData{
				ID:        result.User.ID,
				Email:     result.User.Email,
				CreatedAt: result.User.CreatedAt.UTC().Format(time.RFC3339Nano),
			},
			AccessToken: result.AccessToken,
			TokenType:   result.TokenType,
			ExpiresIn:   result.ExpiresInSec,
		},
	}
}

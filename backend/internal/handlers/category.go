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

const maxCategoryBodyBytes = 1 << 20 // 1 MiB

type CategoryHandler struct {
	service *service.CategoryService
	logger  *slog.Logger
}

func NewCategoryHandler(svc *service.CategoryService, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{service: svc, logger: logger}
}

type categoryResponse struct {
	Data categoryData `json:"data"`
}

type categoryListResponse struct {
	Data []categoryData `json:"data"`
}

type categoryData struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CategoryType string  `json:"category_type"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCategoryBodyBytes)

	var req dto.CreateCategoryRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	category, err := h.service.CreateCategory(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, err, "failed to create category")
		return
	}

	h.logger.Info("category created", "user_id", userID, "category_id", category.ID)
	writeJSON(w, http.StatusCreated, categoryResponse{Data: toCategoryData(category)})
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	categories, err := h.service.ListCategories(r.Context(), userID, r.URL.Query().Get("type"))
	if err != nil {
		h.writeError(w, err, "failed to list categories")
		return
	}

	h.logger.Info("categories listed", "user_id", userID, "count", len(categories))

	data := make([]categoryData, 0, len(categories))
	for i := range categories {
		data = append(data, toCategoryData(&categories[i]))
	}
	writeJSON(w, http.StatusOK, categoryListResponse{Data: data})
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	categoryID := r.PathValue("id")
	if !isPathUUID(categoryID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "category id must be a valid UUID"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCategoryBodyBytes)

	var req dto.UpdateCategoryRequest
	if err := decodeJSONStrict(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	category, err := h.service.UpdateCategory(r.Context(), userID, categoryID, req)
	if err != nil {
		h.writeError(w, err, "failed to update category")
		return
	}

	h.logger.Info("category updated", "user_id", userID, "category_id", category.ID)
	writeJSON(w, http.StatusOK, categoryResponse{Data: toCategoryData(category)})
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	categoryID := r.PathValue("id")
	if !isPathUUID(categoryID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "category id must be a valid UUID"})
		return
	}

	if err := h.service.DeleteCategory(r.Context(), userID, categoryID); err != nil {
		h.writeError(w, err, "failed to delete category")
		return
	}

	h.logger.Info("category deleted", "user_id", userID, "category_id", categoryID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) writeError(w http.ResponseWriter, err error, logMessage string) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrCategoryNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "category not found"})
	case errors.Is(err, domain.ErrCategoryAlreadyExists):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "category already exists"})
	default:
		h.logger.Error(logMessage, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toCategoryData(category *domain.Category) categoryData {
	return categoryData{
		ID:           category.ID,
		Name:         category.Name,
		CategoryType: category.CategoryType,
		Color:        category.Color,
		Icon:         category.Icon,
		CreatedAt:    category.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:    category.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/auth"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/service"
)

type ImportHandler struct {
	service     *service.ImportService
	logger      *slog.Logger
	maxFileSize int64
}

func NewImportHandler(svc *service.ImportService, logger *slog.Logger, maxFileSize int64) *ImportHandler {
	return &ImportHandler{
		service:     svc,
		logger:      logger,
		maxFileSize: maxFileSize,
	}
}

type importJobResponse struct {
	Data importJobData `json:"data"`
}

type importJobListResponse struct {
	Data       []importJobData    `json:"data"`
	Pagination paginationResponse `json:"pagination"`
}

type importJobData struct {
	ID               string  `json:"id"`
	AccountID        string  `json:"account_id"`
	OriginalFilename string  `json:"original_filename"`
	Status           string  `json:"status"`
	TotalRows        int     `json:"total_rows"`
	ProcessedRows    int     `json:"processed_rows"`
	SuccessfulRows   int     `json:"successful_rows"`
	FailedRows       int     `json:"failed_rows"`
	ErrorMessage     *string `json:"error_message"`
	StartedAt        *string `json:"started_at"`
	CompletedAt      *string `json:"completed_at"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type importErrorsResponse struct {
	Data       []importErrorData  `json:"data"`
	Pagination paginationResponse `json:"pagination"`
}

type importErrorData struct {
	RowNumber    int               `json:"row_number"`
	ErrorMessage string            `json:"error_message"`
	RawRow       map[string]string `json:"raw_row"`
}

func (h *ImportHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileSize+1<<20)
	if err := r.ParseMultipartForm(h.maxFileSize + 1<<20); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid multipart form"})
		return
	}

	accountID := strings.TrimSpace(r.FormValue("account_id"))
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	job, err := h.service.CreateTransactionImport(
		r.Context(),
		userID,
		accountID,
		filename,
		file,
		header.Size,
	)
	if err != nil {
		h.writeCreateError(w, err)
		return
	}

	h.logger.Info(
		"import job queued",
		"user_id", userID,
		"job_id", job.ID,
		"account_id", job.AccountID,
		"status", job.Status,
	)

	writeJSON(w, http.StatusAccepted, importJobResponse{Data: toImportJobData(job)})
}

func (h *ImportHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	result, err := h.service.ListImports(
		r.Context(),
		userID,
		r.URL.Query().Get("page"),
		r.URL.Query().Get("page_size"),
	)
	if err != nil {
		h.writeReadError(w, err, "failed to list imports")
		return
	}

	data := make([]importJobData, 0, len(result.Jobs))
	for i := range result.Jobs {
		data = append(data, toImportJobData(&result.Jobs[i]))
	}

	h.logger.Info("imports listed", "user_id", userID, "count", len(data))
	writeJSON(w, http.StatusOK, importJobListResponse{
		Data: data,
		Pagination: paginationResponse{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.TotalItems,
			TotalPages: result.TotalPages,
		},
	})
}

func (h *ImportHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	jobID := r.PathValue("id")
	if !isPathUUID(jobID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "import id must be a valid UUID"})
		return
	}

	job, err := h.service.GetImport(r.Context(), userID, jobID)
	if err != nil {
		h.writeReadError(w, err, "failed to get import")
		return
	}

	h.logger.Info("import retrieved", "user_id", userID, "job_id", job.ID, "status", job.Status)
	writeJSON(w, http.StatusOK, importJobResponse{Data: toImportJobData(job)})
}

func (h *ImportHandler) ListErrors(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	jobID := r.PathValue("id")
	if !isPathUUID(jobID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "import id must be a valid UUID"})
		return
	}

	result, err := h.service.ListImportErrors(
		r.Context(),
		userID,
		jobID,
		r.URL.Query().Get("page"),
		r.URL.Query().Get("page_size"),
	)
	if err != nil {
		h.writeReadError(w, err, "failed to list import errors")
		return
	}

	data := make([]importErrorData, 0, len(result.Errors))
	for _, item := range result.Errors {
		raw := item.RawRow
		if raw == nil {
			raw = map[string]string{}
		}
		data = append(data, importErrorData{
			RowNumber:    item.RowNumber,
			ErrorMessage: item.ErrorMessage,
			RawRow:       raw,
		})
	}

	h.logger.Info("import errors listed", "user_id", userID, "job_id", jobID, "count", len(data))
	writeJSON(w, http.StatusOK, importErrorsResponse{
		Data: data,
		Pagination: paginationResponse{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.TotalItems,
			TotalPages: result.TotalPages,
		},
	})
}

func (h *ImportHandler) writeCreateError(w http.ResponseWriter, err error) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrAccountNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "account not found"})
	case errors.Is(err, domain.ErrImportFileTooLarge):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "file exceeds maximum size"})
	case errors.Is(err, domain.ErrImportQueueUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "import queue unavailable"})
	default:
		h.logger.Error("failed to create import", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func (h *ImportHandler) writeReadError(w http.ResponseWriter, err error, logMessage string) {
	var validationErr *domain.ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: validationErr.Message})
	case errors.Is(err, domain.ErrImportNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "import not found"})
	default:
		h.logger.Error(logMessage, "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func toImportJobData(job *domain.TransactionImportJob) importJobData {
	data := importJobData{
		ID:               job.ID,
		AccountID:        job.AccountID,
		OriginalFilename: job.OriginalFilename,
		Status:           job.Status,
		TotalRows:        job.TotalRows,
		ProcessedRows:    job.ProcessedRows,
		SuccessfulRows:   job.SuccessfulRows,
		FailedRows:       job.FailedRows,
		ErrorMessage:     job.ErrorMessage,
		CreatedAt:        job.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:        job.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if job.StartedAt != nil {
		value := job.StartedAt.UTC().Format(time.RFC3339Nano)
		data.StartedAt = &value
	}
	if job.CompletedAt != nil {
		value := job.CompletedAt.UTC().Format(time.RFC3339Nano)
		data.CompletedAt = &value
	}
	return data
}

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type importJobRepository interface {
	CreateJob(ctx context.Context, input domain.CreateTransactionImportInput) (*domain.TransactionImportJob, error)
	GetJobByID(ctx context.Context, userID, jobID string) (*domain.TransactionImportJob, error)
	ListJobs(ctx context.Context, filter domain.ListImportJobsFilter) ([]domain.TransactionImportJob, int64, error)
	ListErrors(ctx context.Context, filter domain.ListImportErrorsFilter) ([]domain.TransactionImportError, int64, error)
	FailJob(ctx context.Context, jobID, message string) error
	EnsureAccountOwnedByUser(ctx context.Context, userID, accountID string) error
}

type importQueue interface {
	Enqueue(ctx context.Context, jobID string) error
}

type ImportService struct {
	repo        importJobRepository
	queue       importQueue
	uploadDir   string
	maxFileSize int64
}

type ListImportsResult struct {
	Jobs       []domain.TransactionImportJob
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}

type ListImportErrorsResult struct {
	Errors     []domain.TransactionImportError
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}

func NewImportService(
	repo importJobRepository,
	queue importQueue,
	uploadDir string,
	maxFileSize int64,
) *ImportService {
	return &ImportService{
		repo:        repo,
		queue:       queue,
		uploadDir:   uploadDir,
		maxFileSize: maxFileSize,
	}
}

func (s *ImportService) CreateTransactionImport(
	ctx context.Context,
	userID, accountID, originalFilename string,
	file io.Reader,
	fileSize int64,
) (*domain.TransactionImportJob, error) {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(accountID) {
		return nil, &domain.ValidationError{Message: "account_id must be a valid UUID"}
	}
	if err := s.repo.EnsureAccountOwnedByUser(ctx, userID, accountID); err != nil {
		return nil, err
	}

	originalFilename = filepath.Base(strings.TrimSpace(originalFilename))
	if originalFilename == "" || originalFilename == "." || originalFilename == ".." {
		return nil, &domain.ValidationError{Message: "file is required"}
	}
	if !strings.EqualFold(filepath.Ext(originalFilename), ".csv") {
		return nil, &domain.ValidationError{Message: "file must be a .csv"}
	}
	if fileSize > s.maxFileSize {
		return nil, domain.ErrImportFileTooLarge
	}

	if err := os.MkdirAll(s.uploadDir, 0o750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	storedName, err := randomCSVFilename()
	if err != nil {
		return nil, err
	}
	storedPath, err := safeJoin(s.uploadDir, storedName)
	if err != nil {
		return nil, err
	}

	tempPath := storedPath + ".partial"
	if err := writeUploadFile(tempPath, file, s.maxFileSize); err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}
	if err := os.Rename(tempPath, storedPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("finalize upload: %w", err)
	}

	job, err := s.repo.CreateJob(ctx, domain.CreateTransactionImportInput{
		UserID:           userID,
		AccountID:        accountID,
		OriginalFilename: originalFilename,
		StoredFilePath:   storedPath,
	})
	if err != nil {
		_ = os.Remove(storedPath)
		return nil, err
	}

	if err := s.queue.Enqueue(ctx, job.ID); err != nil {
		_ = s.repo.FailJob(ctx, job.ID, "import queue unavailable")
		_ = os.Remove(storedPath)
		return nil, domain.ErrImportQueueUnavailable
	}

	return job, nil
}

func (s *ImportService) ListImports(
	ctx context.Context,
	userID, pageRaw, pageSizeRaw string,
) (*ListImportsResult, error) {
	userID = strings.TrimSpace(userID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}

	page, pageSize, err := parsePageParams(pageRaw, pageSizeRaw)
	if err != nil {
		return nil, err
	}

	jobs, total, err := s.repo.ListJobs(ctx, domain.ListImportJobsFilter{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &ListImportsResult{
		Jobs:       jobs,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *ImportService) GetImport(ctx context.Context, userID, jobID string) (*domain.TransactionImportJob, error) {
	userID = strings.TrimSpace(userID)
	jobID = strings.TrimSpace(jobID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(jobID) {
		return nil, &domain.ValidationError{Message: "import id must be a valid UUID"}
	}
	return s.repo.GetJobByID(ctx, userID, jobID)
}

func (s *ImportService) ListImportErrors(
	ctx context.Context,
	userID, jobID, pageRaw, pageSizeRaw string,
) (*ListImportErrorsResult, error) {
	userID = strings.TrimSpace(userID)
	jobID = strings.TrimSpace(jobID)
	if !isValidUUID(userID) {
		return nil, &domain.ValidationError{Message: "user_id must be a valid UUID"}
	}
	if !isValidUUID(jobID) {
		return nil, &domain.ValidationError{Message: "import id must be a valid UUID"}
	}

	page, pageSize, err := parsePageParams(pageRaw, pageSizeRaw)
	if err != nil {
		return nil, err
	}

	errs, total, err := s.repo.ListErrors(ctx, domain.ListImportErrorsFilter{
		UserID:   userID,
		JobID:    jobID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &ListImportErrorsResult{
		Errors:     errs,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func parsePageParams(pageRaw, pageSizeRaw string) (int, int, error) {
	page := 1
	pageSize := 20

	if value := strings.TrimSpace(pageRaw); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return 0, 0, &domain.ValidationError{Message: "page must be a positive integer"}
		}
		page = parsed
	}
	if value := strings.TrimSpace(pageSizeRaw); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, 0, &domain.ValidationError{Message: "page_size must be an integer between 1 and 100"}
		}
		pageSize = parsed
	}
	return page, pageSize, nil
}

func writeUploadFile(path string, src io.Reader, maxSize int64) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return fmt.Errorf("create upload file: %w", err)
	}
	defer file.Close()

	limited := io.LimitReader(src, maxSize+1)
	written, err := io.Copy(file, limited)
	if err != nil {
		return fmt.Errorf("write upload file: %w", err)
	}
	if written == 0 {
		return &domain.ValidationError{Message: "file must not be empty"}
	}
	if written > maxSize {
		return domain.ErrImportFileTooLarge
	}
	return nil
}

func randomCSVFilename() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate filename: %w", err)
	}
	return hex.EncodeToString(buf) + ".csv", nil
}

func safeJoin(baseDir, name string) (string, error) {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("resolve upload directory: %w", err)
	}
	joined := filepath.Join(baseAbs, name)
	cleaned := filepath.Clean(joined)
	rel, err := filepath.Rel(baseAbs, cleaned)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", &domain.ValidationError{Message: "invalid upload path"}
	}
	return cleaned, nil
}

var csvAmountPattern = regexp.MustCompile(`^[0-9]{1,12}(\.[0-9]{1,2})?$`)

func validateImportAmount(amount string) error {
	amount = strings.TrimSpace(amount)
	if amount == "" || !csvAmountPattern.MatchString(amount) {
		return &domain.ValidationError{Message: "amount must be a positive decimal with at most 2 places"}
	}
	if isAllZeroDecimal(amount) {
		return &domain.ValidationError{Message: "amount must be greater than zero"}
	}
	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
	"github.com/skp7-fordham/fintrack-coach/backend/internal/importcsv"
)

type importProcessorRepository interface {
	GetJobByIDInternal(ctx context.Context, jobID string) (*domain.TransactionImportJob, error)
	MarkJobProcessing(ctx context.Context, jobID string) (*domain.TransactionImportJob, error)
	IsDemoUser(ctx context.Context, userID string) (bool, error)
	UpdateJobProgress(ctx context.Context, jobID string, totalRows, processedRows, successfulRows, failedRows int) error
	CompleteJob(ctx context.Context, jobID, status string, totalRows, processedRows, successfulRows, failedRows int, errorMessage *string) error
	FailJob(ctx context.Context, jobID, message string) error
	AddRowErrors(ctx context.Context, jobID string, errs []domain.RowValidationError) error
	ListUserCategories(ctx context.Context, userID string) ([]domain.Category, error)
	InsertValidatedTransactions(ctx context.Context, userID, accountID string, rows []domain.ValidatedImportRow) error
}

type ImportProcessor struct {
	repo    importProcessorRepository
	maxRows int
	logger  *slog.Logger
}

func NewImportProcessor(repo importProcessorRepository, maxRows int, logger *slog.Logger) *ImportProcessor {
	return &ImportProcessor{
		repo:    repo,
		maxRows: maxRows,
		logger:  logger,
	}
}

func (p *ImportProcessor) ProcessJob(ctx context.Context, jobID string) {
	start := time.Now()

	job, err := p.repo.MarkJobProcessing(ctx, jobID)
	if err != nil {
		if errors.Is(err, domain.ErrImportAlreadyProcessing) {
			p.logger.Info("skipping non-queued import job", "job_id", jobID)
			return
		}
		p.logger.Error("failed to claim import job", "job_id", jobID, "err", err)
		return
	}

	p.logger.Info("import job processing", "job_id", job.ID, "user_id", job.UserID, "account_id", job.AccountID)

	isDemo, err := p.repo.IsDemoUser(ctx, job.UserID)
	if err != nil {
		p.logger.Error("failed to check import job owner", "job_id", job.ID, "err", err)
		_ = p.repo.FailJob(ctx, job.ID, "import processing failed")
		p.cleanupFile(job.StoredFilePath)
		return
	}
	if isDemo {
		if err := p.repo.FailJob(ctx, job.ID, "demo account is read-only"); err != nil {
			p.logger.Error("failed to stop demo import job", "job_id", job.ID, "err", err)
		}
		p.cleanupFile(job.StoredFilePath)
		p.logger.Info("blocked demo import job", "job_id", job.ID, "user_id", job.UserID)
		return
	}

	if err := p.processClaimedJob(ctx, job); err != nil {
		p.logger.Error("import job failed", "job_id", job.ID, "err", err)
		_ = p.repo.FailJob(ctx, job.ID, "import processing failed")
		p.cleanupFile(job.StoredFilePath)
		return
	}

	p.logger.Info(
		"import job finished",
		"job_id", job.ID,
		"user_id", job.UserID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

func (p *ImportProcessor) processClaimedJob(ctx context.Context, job *domain.TransactionImportJob) error {
	file, err := os.Open(job.StoredFilePath)
	if err != nil {
		_ = p.repo.FailJob(ctx, job.ID, "import file missing")
		return fmt.Errorf("open import file: %w", err)
	}
	defer file.Close()

	parsed, err := importcsv.Parse(file, p.maxRows)
	if err != nil {
		message := "invalid csv"
		if errors.Is(err, domain.ErrInvalidCSV) {
			message = err.Error()
		}
		_ = p.repo.FailJob(ctx, job.ID, message)
		p.cleanupFile(job.StoredFilePath)
		return err
	}
	if len(parsed.Rows) == 0 {
		_ = p.repo.FailJob(ctx, job.ID, "no transaction rows found")
		p.cleanupFile(job.StoredFilePath)
		return nil
	}

	categories, err := p.repo.ListUserCategories(ctx, job.UserID)
	if err != nil {
		return err
	}
	categoryIndex := buildCategoryIndex(categories)

	validRows := make([]domain.ValidatedImportRow, 0, len(parsed.Rows))
	rowErrors := make([]domain.RowValidationError, 0)

	for i, row := range parsed.Rows {
		validated, validationErr := validateParsedImportRow(row, categoryIndex)
		if validationErr != nil {
			rowErrors = append(rowErrors, *validationErr)
		} else {
			validRows = append(validRows, *validated)
		}

		if (i+1)%100 == 0 {
			_ = p.repo.UpdateJobProgress(ctx, job.ID, len(parsed.Rows), i+1, len(validRows), len(rowErrors))
		}
	}

	_ = p.repo.UpdateJobProgress(ctx, job.ID, len(parsed.Rows), len(parsed.Rows), len(validRows), len(rowErrors))

	if len(validRows) > 0 {
		if err := p.repo.InsertValidatedTransactions(ctx, job.UserID, job.AccountID, validRows); err != nil {
			_ = p.repo.FailJob(ctx, job.ID, "failed to insert imported transactions")
			p.cleanupFile(job.StoredFilePath)
			return err
		}
	}

	if err := p.repo.AddRowErrors(ctx, job.ID, rowErrors); err != nil {
		_ = p.repo.FailJob(ctx, job.ID, "failed to save import errors")
		p.cleanupFile(job.StoredFilePath)
		return err
	}

	status, message := finalizeImportStatus(len(validRows), len(rowErrors))
	if err := p.repo.CompleteJob(
		ctx,
		job.ID,
		status,
		len(parsed.Rows),
		len(parsed.Rows),
		len(validRows),
		len(rowErrors),
		message,
	); err != nil {
		return err
	}

	p.cleanupFile(job.StoredFilePath)
	return nil
}

func finalizeImportStatus(successCount, failCount int) (string, *string) {
	switch {
	case successCount > 0 && failCount == 0:
		return domain.ImportStatusCompleted, nil
	case successCount > 0 && failCount > 0:
		return domain.ImportStatusCompletedWithErrors, nil
	default:
		msg := "no valid transaction rows"
		return domain.ImportStatusFailed, &msg
	}
}

type categoryRef struct {
	ID           string
	CategoryType string
}

func buildCategoryIndex(categories []domain.Category) map[string]categoryRef {
	index := make(map[string]categoryRef, len(categories))
	for _, category := range categories {
		key := strings.ToLower(strings.TrimSpace(category.Name))
		index[key] = categoryRef{
			ID:           category.ID,
			CategoryType: category.CategoryType,
		}
	}
	return index
}

func validateParsedImportRow(
	row domain.ParsedImportRow,
	categories map[string]categoryRef,
) (*domain.ValidatedImportRow, *domain.RowValidationError) {
	fail := func(message string) *domain.RowValidationError {
		return &domain.RowValidationError{
			RowNumber:    row.RowNumber,
			ErrorMessage: message,
			RawRow:       row.Raw,
		}
	}

	dateValue := strings.TrimSpace(row.Date)
	txDate, err := ParseTransactionDate(dateValue)
	if err != nil {
		return nil, fail("date must use YYYY-MM-DD")
	}

	description, errMsg := clipRequiredText("description", row.Description, 255)
	if errMsg != "" {
		return nil, fail(errMsg)
	}

	merchant, errMsg := clipOptionalText("merchant", row.Merchant, 150)
	if errMsg != "" {
		return nil, fail(errMsg)
	}

	amount := strings.TrimSpace(row.Amount)
	if err := validateImportAmount(amount); err != nil {
		return nil, fail(err.Error())
	}

	txType := strings.ToLower(strings.TrimSpace(row.Type))
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		return nil, fail("type must be income or expense")
	}

	var categoryID *string
	categoryName := strings.TrimSpace(row.Category)
	if categoryName != "" {
		ref, ok := categories[strings.ToLower(categoryName)]
		if !ok {
			return nil, fail("category does not exist")
		}
		if ref.CategoryType != txType {
			return nil, fail("category type must match transaction type")
		}
		id := ref.ID
		categoryID = &id
	}

	notes, errMsg := clipOptionalText("notes", row.Notes, 2000)
	if errMsg != "" {
		return nil, fail(errMsg)
	}

	validated := &domain.ValidatedImportRow{
		RowNumber:       row.RowNumber,
		TransactionDate: txDate,
		Description:     description,
		Amount:          amount,
		TransactionType: txType,
		CategoryID:      categoryID,
	}
	if merchant != "" {
		validated.Merchant = &merchant
	}
	if notes != "" {
		validated.Notes = &notes
	}
	return validated, nil
}

func clipRequiredText(field, value string, maxLen int) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", field + " is required"
	}
	if utf8.RuneCountInString(value) > maxLen {
		return "", fmt.Sprintf("%s must be at most %d characters", field, maxLen)
	}
	return value, ""
}

func clipOptionalText(field, value string, maxLen int) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if utf8.RuneCountInString(value) > maxLen {
		return "", fmt.Sprintf("%s must be at most %d characters", field, maxLen)
	}
	return value, ""
}

func (p *ImportProcessor) cleanupFile(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		p.logger.Error("failed to remove import file", "err", err)
	}
}

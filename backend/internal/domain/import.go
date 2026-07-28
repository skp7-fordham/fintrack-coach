package domain

import (
	"errors"
	"time"
)

const (
	ImportStatusQueued              = "queued"
	ImportStatusProcessing          = "processing"
	ImportStatusCompleted           = "completed"
	ImportStatusCompletedWithErrors = "completed_with_errors"
	ImportStatusFailed              = "failed"
)

var (
	ErrImportNotFound          = errors.New("import job not found")
	ErrImportFileTooLarge      = errors.New("import file too large")
	ErrInvalidCSV              = errors.New("invalid csv")
	ErrImportQueueUnavailable  = errors.New("import queue unavailable")
	ErrImportAlreadyProcessing = errors.New("import job already processing")
)

// TransactionImportJob is the persisted import job state.
type TransactionImportJob struct {
	ID               string
	UserID           string
	AccountID        string
	OriginalFilename string
	StoredFilePath   string
	Status           string
	TotalRows        int
	ProcessedRows    int
	SuccessfulRows   int
	FailedRows       int
	ErrorMessage     *string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TransactionImportError is one row-level import failure.
type TransactionImportError struct {
	ID           int64
	ImportJobID  string
	RowNumber    int
	ErrorMessage string
	RawRow       map[string]string
	CreatedAt    time.Time
}

type CreateTransactionImportInput struct {
	UserID           string
	AccountID        string
	OriginalFilename string
	StoredFilePath   string
}

type ListImportJobsFilter struct {
	UserID   string
	Page     int
	PageSize int
}

type ListImportErrorsFilter struct {
	UserID   string
	JobID    string
	Page     int
	PageSize int
}

// ParsedImportRow is one CSV data row before domain validation.
type ParsedImportRow struct {
	RowNumber   int
	Date        string
	Description string
	Merchant    string
	Amount      string
	Type        string
	Category    string
	Notes       string
	Raw         map[string]string
}

// ValidatedImportRow is ready for database insertion.
type ValidatedImportRow struct {
	RowNumber       int
	TransactionDate time.Time
	Description     string
	Merchant        *string
	Amount          string
	TransactionType string
	CategoryID      *string
	Notes           *string
}

// RowValidationError captures a failed CSV row.
type RowValidationError struct {
	RowNumber    int
	ErrorMessage string
	RawRow       map[string]string
}

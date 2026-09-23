package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type demoImportProcessorRepository struct {
	job         *domain.TransactionImportJob
	failMessage string
	inserted    bool
}

func (r *demoImportProcessorRepository) GetJobByIDInternal(
	context.Context,
	string,
) (*domain.TransactionImportJob, error) {
	return r.job, nil
}

func (r *demoImportProcessorRepository) MarkJobProcessing(
	context.Context,
	string,
) (*domain.TransactionImportJob, error) {
	return r.job, nil
}

func (r *demoImportProcessorRepository) IsDemoUser(context.Context, string) (bool, error) {
	return true, nil
}

func (r *demoImportProcessorRepository) UpdateJobProgress(
	context.Context,
	string,
	int,
	int,
	int,
	int,
) error {
	return nil
}

func (r *demoImportProcessorRepository) CompleteJob(
	context.Context,
	string,
	string,
	int,
	int,
	int,
	int,
	*string,
) error {
	return nil
}

func (r *demoImportProcessorRepository) FailJob(_ context.Context, _ string, message string) error {
	r.failMessage = message
	return nil
}

func (r *demoImportProcessorRepository) AddRowErrors(
	context.Context,
	string,
	[]domain.RowValidationError,
) error {
	return nil
}

func (r *demoImportProcessorRepository) ListUserCategories(
	context.Context,
	string,
) ([]domain.Category, error) {
	return nil, nil
}

func (r *demoImportProcessorRepository) InsertValidatedTransactions(
	context.Context,
	string,
	string,
	[]domain.ValidatedImportRow,
) error {
	r.inserted = true
	return nil
}

func TestImportProcessorBlocksQueuedDemoJob(t *testing.T) {
	repo := &demoImportProcessorRepository{
		job: &domain.TransactionImportJob{
			ID:             "22222222-2222-2222-2222-222222222222",
			UserID:         "11111111-1111-1111-1111-111111111111",
			AccountID:      "33333333-3333-3333-3333-333333333333",
			StoredFilePath: t.TempDir() + "/already-removed.csv",
		},
	}
	processor := NewImportProcessor(repo, 1000, slog.New(slog.NewTextHandler(io.Discard, nil)))

	processor.ProcessJob(context.Background(), repo.job.ID)

	if repo.inserted {
		t.Fatal("demo import inserted transactions")
	}
	if repo.failMessage != "demo account is read-only" {
		t.Fatalf("failure message = %q", repo.failMessage)
	}
}

package importworker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// JobQueue is the Redis import queue consumed by workers.
type JobQueue interface {
	Dequeue(ctx context.Context) (string, error)
}

// Processor runs a claimed import job to completion.
type Processor interface {
	ProcessJob(ctx context.Context, jobID string)
}

// Start launches concurrency worker goroutines that consume import jobs until ctx is cancelled.
// Worker errors and panics are logged and do not stop the other goroutines.
func Start(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs JobQueue,
	processor Processor,
	concurrency int,
	logger *slog.Logger,
) {
	if concurrency < 1 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			run(ctx, workerID, jobs, processor, logger)
		}(i + 1)
	}
}

func run(ctx context.Context, workerID int, jobs JobQueue, processor Processor, logger *slog.Logger) {
	logger.Info("import worker started", "worker_id", workerID)
	for {
		jobID, err := jobs.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, redis.ErrClosed) {
				logger.Info("import worker stopping", "worker_id", workerID)
				return
			}
			logger.Error("failed to dequeue import job", "worker_id", workerID, "err", err)
			select {
			case <-ctx.Done():
				logger.Info("import worker stopping", "worker_id", workerID)
				return
			case <-time.After(time.Second):
			}
			continue
		}

		processSafely(ctx, workerID, jobID, processor, logger)
	}
}

func processSafely(ctx context.Context, workerID int, jobID string, processor Processor, logger *slog.Logger) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("import worker panic recovered", "worker_id", workerID, "job_id", jobID, "panic", rec)
		}
	}()
	processor.ProcessJob(ctx, jobID)
}

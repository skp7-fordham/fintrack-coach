package importworker

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeQueue struct {
	jobs chan string
}

func (q *fakeQueue) Dequeue(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case jobID := <-q.jobs:
		return jobID, nil
	}
}

type fakeProcessor struct {
	count atomic.Int32
}

func (p *fakeProcessor) ProcessJob(ctx context.Context, jobID string) {
	p.count.Add(1)
}

func TestStartProcessesJobAndStops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	q := &fakeQueue{jobs: make(chan string, 1)}
	p := &fakeProcessor{}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	Start(ctx, &wg, q, p, 1, logger)

	q.jobs <- "job-1"

	deadline := time.Now().Add(2 * time.Second)
	for p.count.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if p.count.Load() != 1 {
		t.Fatalf("processed = %d, want 1", p.count.Load())
	}

	cancel()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not stop")
	}
}

type panicProcessor struct{}

func (panicProcessor) ProcessJob(ctx context.Context, jobID string) {
	panic("boom")
}

func TestStartRecoversFromProcessorPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	q := &fakeQueue{jobs: make(chan string, 1)}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	Start(ctx, &wg, q, panicProcessor{}, 1, logger)

	q.jobs <- "job-panic"
	time.Sleep(50 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("panic in worker must not hang shutdown")
	}
}

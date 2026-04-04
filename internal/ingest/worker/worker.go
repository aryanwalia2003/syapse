package worker

import (
	"context"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/errors"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

const maxRetries = 3

type WorkerConfig struct {
	Repo domain.WebhookRepository
	Name string
}

type Worker struct {
	repo domain.WebhookRepository
	name string
}

func NewWorker(config WorkerConfig) *Worker {
	return &Worker{
		repo: config.Repo,
		name: config.Name,
	}
}

// targetFunc abstracts the actual normalizer execution to allow easier testing.
// In actual use, this function invokes the normalization engine.
func (w *Worker) processWebhook(ctx context.Context, wh *domain.RawWebhook, targetFunc func(context.Context, *domain.RawWebhook) error) error {
	var err error

	// Establish a func to capture panics during targetFunc execution
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(errors.CodeInternal, "worker panic recovered")
				logger.Error(ctx, "worker panic recovered", err, "panic", r)
			}
		}()
		err = targetFunc(ctx, wh)
	}()

	// Handle context timeout - assume failure but don't increment retry. The sweeper will catch it later.
	if err == context.DeadlineExceeded || ctx.Err() != nil {
		logger.Error(ctx, "context timed out or cancelled during processing", err, "id", wh.ID)
		return err
	}

	if err != nil {
		logger.Error(ctx, "processing failed", err, "id", wh.ID, "retry_count", wh.RetryCount)
		return w.handleFailure(ctx, wh)
	}

	return w.repo.MarkDone(ctx, wh.ID)
}

func (w *Worker) handleFailure(ctx context.Context, wh *domain.RawWebhook) error {
	w.repo.IncrementRetry(ctx, wh.ID)
	// After IncrementRetry, the RetryCount is technically wh.RetryCount + 1.
	// But we check if it has already reached the maximum BEFORE this failure!
	// Wait, actually testing maxRetries = 3.
	// If it was 0, it becomes 1. If 1, becomes 2. If 2, becomes 3.
	// If it was 3 previously (maybe it failed 3 times), we mark it failed.
	// Or we mark it failed if retry count will exceed max. Check PRD.
	if wh.RetryCount >= maxRetries {
		logger.Error(ctx, "max retries exceeded, moving to DLQ", nil, "id", wh.ID)
		w.repo.MarkFailed(ctx, wh.ID)
	}
	return nil
}

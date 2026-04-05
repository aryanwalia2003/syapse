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

// Run reads webhooks from ch and calls processWebhook for each one.
// It exits when ctx is cancelled or ch is closed.
func (w *Worker) Run(ctx context.Context, ch <-chan *domain.RawWebhook, targetFunc func(context.Context, *domain.RawWebhook) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case wh, ok := <-ch:
			if !ok {
				return
			}
			if err := w.processWebhook(ctx, wh, targetFunc); err != nil {
				logger.Error(ctx, "worker failed to process webhook", err, "worker", w.name, "id", wh.ID)
			}
		}
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
	if wh.RetryCount+1 >= maxRetries {
		w.repo.MarkFailed(ctx, wh.ID)
		logger.Error(ctx, "max retries exceeded, moving to DLQ", nil, "id", wh.ID)
		return nil
	}
	w.repo.IncrementRetry(ctx, wh.ID)
	return nil
}

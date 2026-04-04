package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"log/slog"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func init() {
	logger.Init(slog.LevelDebug, false)
}

type mockRepo struct {
	domain.WebhookRepository
	incRetryCount int
	failedCalled  bool
	doneCalled    bool
	lastStatus    string
}

func (m *mockRepo) IncrementRetry(ctx context.Context, id string) error {
	m.incRetryCount++
	return nil
}

func (m *mockRepo) MarkFailed(ctx context.Context, id string) error {
	m.failedCalled = true
	m.lastStatus = "FAILED"
	return nil
}

func (m *mockRepo) MarkDone(ctx context.Context, id string) error {
	m.doneCalled = true
	m.lastStatus = "DONE"
	return nil
}

func TestWorkerRetryLogic(t *testing.T) {
	t.Run("normalizer success marks done", func(t *testing.T) {
		repo := &mockRepo{}
		w := NewWorker(WorkerConfig{Repo: repo, Name: "w1"})

		wh := &domain.RawWebhook{ID: "wh-1"}

		// passing normalizer
		err := w.processWebhook(context.Background(), wh, func(ctx context.Context, wh *domain.RawWebhook) error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !repo.doneCalled {
			t.Errorf("Expected MarkDone to be called")
		}
	})

	t.Run("normalizer error increments retry", func(t *testing.T) {
		repo := &mockRepo{}
		w := NewWorker(WorkerConfig{Repo: repo, Name: "w1"})

		wh := &domain.RawWebhook{ID: "wh-1", RetryCount: 0}

		// failing normalizer
		w.processWebhook(context.Background(), wh, func(ctx context.Context, wh *domain.RawWebhook) error {
			return errors.New("norm error")
		})

		if repo.incRetryCount != 1 {
			t.Errorf("Expected IncrementRetry to be called once, got %d", repo.incRetryCount)
		}
		if repo.failedCalled {
			t.Errorf("Expected MarkFailed not to be called yet")
		}
	})

	t.Run("normalizer panic is caught and increments retry", func(t *testing.T) {
		repo := &mockRepo{}
		w := NewWorker(WorkerConfig{Repo: repo, Name: "w1"})

		wh := &domain.RawWebhook{ID: "wh-1", RetryCount: 0}

		// panicking normalizer
		w.processWebhook(context.Background(), wh, func(ctx context.Context, wh *domain.RawWebhook) error {
			panic("nil pointer")
		})

		if repo.incRetryCount != 1 {
			t.Errorf("Expected panic to trigger IncrementRetry")
		}
	})

	t.Run("third failure transitions to DLQ", func(t *testing.T) {
		repo := &mockRepo{}
		w := NewWorker(WorkerConfig{Repo: repo, Name: "w1"})

		wh := &domain.RawWebhook{ID: "wh-1", RetryCount: 3} // Already retried 3 times (0, 1, 2, 3 means this is 4th attempt or exceeded max)
		// Or let's say maxRetries is 3. If RetryCount == 2 and fails, it increments to 3 and marks failed.

		w.processWebhook(context.Background(), wh, func(ctx context.Context, wh *domain.RawWebhook) error {
			return errors.New("norm error")
		})

		if !repo.failedCalled {
			t.Errorf("Expected MarkFailed to be called when max retries reached")
		}
	})

	t.Run("context timeout aborts without retry increment", func(t *testing.T) {
		repo := &mockRepo{}
		w := NewWorker(WorkerConfig{Repo: repo, Name: "w1"})

		wh := &domain.RawWebhook{ID: "wh-1", RetryCount: 0}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// slow normalizer
		w.processWebhook(ctx, wh, func(ctx context.Context, wh *domain.RawWebhook) error {
			<-ctx.Done()
			return ctx.Err()
		})

		if repo.incRetryCount > 0 {
			t.Errorf("Expected timeout NOT to increment retry count (should be recovered by sweeper)")
		}
	})
}

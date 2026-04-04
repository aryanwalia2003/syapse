package sweeper

import (
	"context"
	"testing"
	"time"

	"log/slog"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func init() {
	logger.Init(slog.LevelDebug, false)
}

type mockSweeperRepo struct {
	domain.WebhookRepository
	recovered int64
}

func (m *mockSweeperRepo) RecoverStuck(ctx context.Context, threshold time.Duration) (int64, error) {
	return m.recovered, nil
}

func TestSweeper(t *testing.T) {
	t.Run("sweeper recovers stuck jobs", func(t *testing.T) {
		repo := &mockSweeperRepo{recovered: 5}

		sw := NewSweeper(repo, 10*time.Minute)

		// Run one sweep manually
		recovered, err := sw.sweep(context.Background())

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if recovered != 5 {
			t.Errorf("Expected 5 recovered jobs, got %d", recovered)
		}
	})

	t.Run("sweeper stops cleanly on cancel", func(t *testing.T) {
		repo := &mockSweeperRepo{recovered: 0}
		sw := NewSweeper(repo, 1*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			sw.Run(ctx)
			close(done)
		}()

		cancel() // Should immediately stop Run()

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Errorf("Sweeper did not stop gracefully on context cancel")
		}
	})
}

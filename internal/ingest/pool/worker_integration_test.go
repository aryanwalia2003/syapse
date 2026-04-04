package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	"log/slog"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func init() {
	logger.Init(slog.LevelDebug, false)
}

// A simple mock repo for integration tests
type mockRepo struct {
	domain.WebhookRepository
}

func (m *mockRepo) MarkDone(ctx context.Context, id string) error {
	return nil
}

func TestWorkerIntegration_SequentialProcessing(t *testing.T) {
	// The PRD mentions: "sequential processing for same OrderID"

	// We'll capture the execution order
	var mu sync.Mutex
	var executionOrder []string

	targetFunc := func(ctx context.Context, wh *domain.RawWebhook) error {
		// simulate some processing time
		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		executionOrder = append(executionOrder, wh.ID)
		mu.Unlock()

		return nil
	}

	repo := &mockRepo{}
	pool := NewWorkerPool(8, repo, targetFunc)

	// manually start workers for testing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	startWorker := func(wChan chan *domain.RawWebhook) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case wh, ok := <-wChan:
					if !ok {
						return
					}
					// For testing, just call targetFunc directly
					targetFunc(ctx, wh)
				}
			}
		}()
	}

	for _, ch := range pool.workerChans {
		startWorker(ch)
	}

	// We send 3 webhooks all to the same partition
	partIdx := 5
	pool.Route(&domain.RawWebhook{ID: "wh-1", PartitionIndex: partIdx})
	pool.Route(&domain.RawWebhook{ID: "wh-2", PartitionIndex: partIdx})
	pool.Route(&domain.RawWebhook{ID: "wh-3", PartitionIndex: partIdx})

	// Wait for them to be processed
	time.Sleep(250 * time.Millisecond)

	pool.Stop()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(executionOrder) != 3 {
		t.Fatalf("Expected 3 executions, got %d", len(executionOrder))
	}
	if executionOrder[0] != "wh-1" || executionOrder[1] != "wh-2" || executionOrder[2] != "wh-3" {
		t.Errorf("Expected sequential execution [wh-1 wh-2 wh-3], got %v", executionOrder)
	}
}

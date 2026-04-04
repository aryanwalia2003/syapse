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

type pollingMockRepo struct {
	domain.WebhookRepository
	mu                 sync.Mutex
	pendingByPartition map[int][]*domain.RawWebhook
	markProcessingIDs  []string
}

func (m *pollingMockRepo) ClaimPending(ctx context.Context, partitionIndex int, limit int) ([]*domain.RawWebhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return once — drain so we don't keep re-dispatching in subsequent polls
	results := m.pendingByPartition[partitionIndex]
	m.pendingByPartition[partitionIndex] = nil
	return results, nil
}

func (m *pollingMockRepo) MarkProcessing(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markProcessingIDs = append(m.markProcessingIDs, id)
	return nil
}

func TestPollingLoop_DispatchesToCorrectChannel(t *testing.T) {
	repo := &pollingMockRepo{
		pendingByPartition: map[int][]*domain.RawWebhook{
			2: {{ID: "wh-1", PartitionIndex: 2}, {ID: "wh-2", PartitionIndex: 2}},
			5: {{ID: "wh-3", PartitionIndex: 5}},
		},
	}

	targetFunc := func(ctx context.Context, wh *domain.RawWebhook) error { return nil }
	pool := NewWorkerPool(PoolConfig{N: 8, PollInterval: 10 * time.Millisecond, ClaimLimit: 20}, repo, targetFunc)

	ctx, cancel := context.WithCancel(context.Background())

	received := make(chan string, 10)

	// Watch channels 2 and 5 to validate dispatch
	go func() {
		for wh := range pool.workerChans[2] {
			received <- wh.ID
		}
	}()
	go func() {
		for wh := range pool.workerChans[5] {
			received <- wh.ID
		}
	}()

	pool.Start(ctx)
	time.Sleep(150 * time.Millisecond) // let first poll cycle complete
	cancel()
	pool.wg.Wait() // wait for all pollers to exit cleanly

	repo.mu.Lock()
	processed := len(repo.markProcessingIDs)
	repo.mu.Unlock()

	if processed == 0 {
		t.Errorf("Expected MarkProcessing to be called for dispatched jobs")
	}
}

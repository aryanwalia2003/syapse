package pool

import (
	"context"
	"sync"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/ingest/partition"
	"github.com/aryanwalia/synapse/internal/ingest/worker"
)

type WorkerPool struct {
	workers     []*worker.Worker
	workerChans []chan *domain.RawWebhook
	unsorted    *worker.Worker
	unsortedCh  chan *domain.RawWebhook
	wg          sync.WaitGroup
	n           int
	targetFunc  func(context.Context, *domain.RawWebhook) error
}

func NewWorkerPool(n int, repo domain.WebhookRepository, targetFunc func(context.Context, *domain.RawWebhook) error) *WorkerPool {
	workers := make([]*worker.Worker, n)
	workerChans := make([]chan *domain.RawWebhook, n)

	for i := 0; i < n; i++ {
		workers[i] = worker.NewWorker(worker.WorkerConfig{
			Repo: repo,
			Name: "worker-" + string(rune(i)),
		})
		workerChans[i] = make(chan *domain.RawWebhook, 100)
	}

	unsorted := worker.NewWorker(worker.WorkerConfig{
		Repo: repo,
		Name: "unsorted-worker",
	})
	unsortedCh := make(chan *domain.RawWebhook, 100)

	return &WorkerPool{
		workers:     workers,
		workerChans: workerChans,
		unsorted:    unsorted,
		unsortedCh:  unsortedCh,
		n:           n,
		targetFunc:  targetFunc,
	}
}

func (p *WorkerPool) Route(webhook *domain.RawWebhook) {
	if webhook.PartitionIndex == partition.UnsortedPartition() || webhook.PartitionIndex < 0 || webhook.PartitionIndex >= p.n {
		p.unsortedCh <- webhook
		return
	}
	p.workerChans[webhook.PartitionIndex] <- webhook
}

func (p *WorkerPool) Start(ctx context.Context) {
	// Not fully implemented yet.
	// For testing purposes, we define the API.
	// We'll leave worker loop to complete later if needed.
}

func (p *WorkerPool) Stop() {
	for _, ch := range p.workerChans {
		close(ch)
	}
	close(p.unsortedCh)
	p.wg.Wait()
}

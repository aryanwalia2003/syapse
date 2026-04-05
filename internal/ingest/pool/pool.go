package pool

import (
	"context"
	"sync"
	"time"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
	"github.com/aryanwalia/synapse/internal/ingest/partition"
	"github.com/aryanwalia/synapse/internal/ingest/worker"
)

const (
	defaultPollInterval = 500 * time.Millisecond
	defaultClaimLimit   = 20
)

type PoolConfig struct {
	N            int
	PollInterval time.Duration // default: 500ms
	ClaimLimit   int           // default: 20
}

func (c *PoolConfig) withDefaults() PoolConfig {
	if c.PollInterval == 0 {
		c.PollInterval = defaultPollInterval
	}
	if c.ClaimLimit == 0 {
		c.ClaimLimit = defaultClaimLimit
	}
	return *c
}

type WorkerPool struct {
	repo        domain.WebhookRepository
	workers     []*worker.Worker
	workerChans []chan *domain.RawWebhook
	unsorted    *worker.Worker
	unsortedCh  chan *domain.RawWebhook
	wg          sync.WaitGroup
	cfg         PoolConfig
	targetFunc  func(context.Context, *domain.RawWebhook) error
}

func NewWorkerPool(cfg PoolConfig, repo domain.WebhookRepository, targetFunc func(context.Context, *domain.RawWebhook) error) *WorkerPool {
	cfg = cfg.withDefaults()
	n := cfg.N
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

	return &WorkerPool{
		repo:        repo,
		workers:     workers,
		workerChans: workerChans,
		unsorted:    unsorted,
		unsortedCh:  make(chan *domain.RawWebhook, 100),
		cfg:         cfg,
		targetFunc:  targetFunc,
	}
}

func (p *WorkerPool) Route(webhook *domain.RawWebhook) {
	if webhook.PartitionIndex == partition.UnsortedPartition() || webhook.PartitionIndex < 0 || webhook.PartitionIndex >= p.cfg.N {
		p.unsortedCh <- webhook
		return
	}
	p.workerChans[webhook.PartitionIndex] <- webhook
}

// Start launches N partition pollers + 1 unsorted poller, each running in their own goroutine.
// It also launches a consumer goroutine per worker channel so that dispatched webhooks are
// actually processed — without these the channels would fill up and block forever.
func (p *WorkerPool) Start(ctx context.Context) {
	// Launch partition pollers.
	for i := 0; i < p.cfg.N; i++ {
		p.wg.Add(1)
		go p.runPoller(ctx, i)
	}
	// Launch unsorted poller.
	p.wg.Add(1)
	go p.runPoller(ctx, partition.UnsortedPartition())

	// Launch consumers — one goroutine per worker channel.
	for i := 0; i < p.cfg.N; i++ {
		w := p.workers[i]
		ch := p.workerChans[i]
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			w.Run(ctx, ch, p.targetFunc)
		}()
	}
	// Launch consumer for the unsorted channel.
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.unsorted.Run(ctx, p.unsortedCh, p.targetFunc)
	}()
}

func (p *WorkerPool) runPoller(ctx context.Context, partitionIdx int) {
	defer p.wg.Done()
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.claimAndDispatch(ctx, partitionIdx)
		}
	}
}

func (p *WorkerPool) claimAndDispatch(ctx context.Context, partitionIdx int) {
	webhooks, err := p.repo.ClaimPending(ctx, partitionIdx, p.cfg.ClaimLimit)
	if err != nil {
		logger.Error(ctx, "failed to claim pending webhooks", err, "partition", partitionIdx)
		return
	}
	for _, wh := range webhooks {
		if markErr := p.repo.MarkProcessing(ctx, wh.ID); markErr != nil {
			logger.Error(ctx, "failed to mark webhook as processing", markErr, "id", wh.ID)
			continue
		}
		p.Route(wh)
	}
}

func (p *WorkerPool) Stop() {
	for _, ch := range p.workerChans {
		close(ch)
	}
	close(p.unsortedCh)
	p.wg.Wait()
}

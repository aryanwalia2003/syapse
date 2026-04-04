package sweeper

import (
	"context"
	"time"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

type Sweeper struct {
	repo     domain.WebhookRepository
	interval time.Duration
}

func NewSweeper(repo domain.WebhookRepository, interval time.Duration) *Sweeper {
	return &Sweeper{repo: repo, interval: interval}
}

func (s *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sweep(ctx)
		}
	}
}

func (s *Sweeper) sweep(ctx context.Context) (int64, error) {
	recovered, err := s.repo.RecoverStuck(ctx, s.interval)
	if err != nil {
		logger.Error(ctx, "failed to sweep stuck jobs", err)
		return 0, err
	}
	if recovered > 0 {
		logger.Info(ctx, "swept stuck jobs", "count", recovered)
	}
	return recovered, nil
}

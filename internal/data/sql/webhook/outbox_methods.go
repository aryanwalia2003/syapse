package webhook

import (
	"context"
	"time"

	"github.com/aryanwalia/synapse/internal/core/domain"
)

func (repo *SQLWebhookRepository) ClaimPending(ctx context.Context, partitionIndex int, limit int) ([]*domain.RawWebhook, error) {
	return nil, nil // TODO
}

func (repo *SQLWebhookRepository) MarkProcessing(ctx context.Context, id string) error {
	return nil // TODO
}

func (repo *SQLWebhookRepository) MarkDone(ctx context.Context, id string) error {
	return nil // TODO
}

func (repo *SQLWebhookRepository) MarkFailed(ctx context.Context, id string) error {
	return nil // TODO
}

func (repo *SQLWebhookRepository) IncrementRetry(ctx context.Context, id string) error {
	return nil // TODO
}

func (repo *SQLWebhookRepository) RecoverStuck(ctx context.Context, stuckThreshold time.Duration) (int64, error) {
	return 0, nil // TODO
}

package webhook

import (
	"context"
	"time"

	"github.com/aryanwalia/synapse/internal/core/domain"
)

// ClaimPending fetches PENDING webhooks for a specific partition index.
func (repo *SQLWebhookRepository) ClaimPending(ctx context.Context, partitionIndex int, limit int) ([]*domain.RawWebhook, error) {
	sql := `
		SELECT id, status, retry_count, partition_index
		FROM raw_webhook_payloads
		WHERE status = 'PENDING' AND partition_index = {:partition_index}
		LIMIT {:limit}
	`
	var results []*domain.RawWebhook
	err := repo.db.QueryRows(ctx, sql, map[string]any{
		"partition_index": partitionIndex,
		"limit":           limit,
	}, &results)
	return results, err
}

// MarkProcessing transitions a webhook row to PROCESSING state.
func (repo *SQLWebhookRepository) MarkProcessing(ctx context.Context, id string) error {
	sql := `UPDATE raw_webhook_payloads SET status = {:status}, updated = {:updated} WHERE id = {:id}`
	return repo.db.Execute(ctx, sql, map[string]any{
		"id":      id,
		"status":  "PROCESSING",
		"updated": time.Now().UTC(),
	})
}

// MarkDone transitions a webhook row to DONE state.
func (repo *SQLWebhookRepository) MarkDone(ctx context.Context, id string) error {
	sql := `UPDATE raw_webhook_payloads SET status = {:status}, updated = {:updated} WHERE id = {:id}`
	return repo.db.Execute(ctx, sql, map[string]any{
		"id":      id,
		"status":  "DONE",
		"updated": time.Now().UTC(),
	})
}

// MarkFailed transitions a webhook to the FAILED/DLQ state after max retries.
func (repo *SQLWebhookRepository) MarkFailed(ctx context.Context, id string) error {
	sql := `
		UPDATE raw_webhook_payloads
		SET status = {:status}, is_dlq = {:is_dlq}, updated = {:updated}
		WHERE id = {:id}
	`
	return repo.db.Execute(ctx, sql, map[string]any{
		"id":      id,
		"status":  "FAILED",
		"is_dlq":  true,
		"updated": time.Now().UTC(),
	})
}

// IncrementRetry increments the retry_count for a specific webhook.
func (repo *SQLWebhookRepository) IncrementRetry(ctx context.Context, id string) error {
	sql := `
		UPDATE raw_webhook_payloads
		SET retry_count = retry_count + 1, updated = {:updated}
		WHERE id = {:id}
	`
	return repo.db.Execute(ctx, sql, map[string]any{
		"id":      id,
		"updated": time.Now().UTC(),
	})
}

// RecoverStuck resets PROCESSING webhooks stuck longer than the threshold back to PENDING.
func (repo *SQLWebhookRepository) RecoverStuck(ctx context.Context, stuckThreshold time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-stuckThreshold)
	sql := `
		UPDATE raw_webhook_payloads
		SET status = 'PENDING', updated = {:updated}
		WHERE status = 'PROCESSING' AND updated < {:cutoff}
	`
	err := repo.db.Execute(ctx, sql, map[string]any{
		"cutoff":  cutoff,
		"updated": time.Now().UTC(),
	})
	// PocketBase/SQLite doesn't easily return affected rows via this interface.
	// In production Postgres, we'd use RETURNING or a SELECT COUNT first.
	return 0, err
}

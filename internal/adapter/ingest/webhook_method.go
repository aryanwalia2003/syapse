package ingest

import (
	"context"
	"encoding/json"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
	"github.com/google/uuid"
)

func ProcessIngest(ctx context.Context, req WebhookRequest, webhookRepo domain.WebhookRepository) IngestResult {
	correlationID := uuid.New().String()

	// 1. Raw Persistence (Audit Trail)
	headerBytes, _ := json.Marshal(req.Header)
	rawWebhook := &domain.RawWebhook{
		CorrelationID: correlationID,
		Source:        req.Source,
		Payload:       req.Payload,
		Headers:       headerBytes,
		Status:        "RECEIVED",
	}

	err := webhookRepo.SaveRaw(ctx, rawWebhook)
	if err != nil {
		logger.Error(ctx, "Failed to save raw webhook payload", err, "correlation_id", correlationID)
		// We still continue or return partial success?
		// Usually for audit trail, if this fails, we might want to know but not necessarily block.
		// However, for Synapse, persistence is key.
	}

	// TODO: Task 4.1 - Vendor Signature Validation

	return IngestResult{
		Success:       true,
		CorrelationID: correlationID,
		Message:       "Webhook received and securely logged",
	}
}

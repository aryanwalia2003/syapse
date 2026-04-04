package ingest

import (
	"context"
	"encoding/json"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/errors"
	"github.com/aryanwalia/synapse/internal/core/logger"
	"github.com/google/uuid"
)

func (p *IngestPipeline) Process(ctx context.Context, req WebhookRequest) IngestResult {
	correlationID := uuid.New().String()

	if persistErr := p.persistRawWebhook(ctx, correlationID, req); persistErr != nil {
		logger.Error(ctx, "audit trail persistence failed", persistErr,
			"correlation_id", correlationID,
			"provider", req.ProviderName,
		)

		return IngestResult{
			Success:       false,
			CorrelationID: correlationID,
			Message:       "Internal Server Error ; failed to securely persist raw webhook.",
		}
	}

	logger.Info(ctx, "webhook persisted successfully",
		"correlation_id", correlationID,
		"provider", req.ProviderName,
		"webhook_type", req.Type,
	)

	return IngestResult{
		Success:       true,
		CorrelationID: correlationID,
		Message:       "Webhook accepted.",
	}
}

func (p *IngestPipeline) persistRawWebhook(ctx context.Context, correlationID string, req WebhookRequest) error {
	headerBytes, err := json.Marshal(req.Header)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternal, "failed to marshal request headers for audit trail")
	}

	// Just hard-code PENDING outbox fields - extracting logic can be updated later if needed.
	// Assume partition and extractor logic applies in worker layer or here, we are making it bare minimal.
	rawWebhook := &domain.RawWebhook{
		CorrelationID: correlationID,
		Source:        req.ProviderName,
		Payload:       req.Payload,
		Headers:       headerBytes,
		Status:        "PENDING", // Changed from RECEIVED to PENDING
		ReceivedAt:    req.ReceivedAt,
	}

	return p.webhookRepository.SaveRaw(ctx, rawWebhook)
}
func (p *IngestPipeline) normalize(ctx context.Context, correlationID string, req WebhookRequest) error {
	switch req.Type {
	case WebhookTypeDISStatusUpdate:
		return p.normalizeDISStatusUpdate(ctx, correlationID, req)

	case WebhookTypeWMSOrderCreation:
		return p.normalizeWMSOrderCreation(ctx, correlationID, req)

	default:
		return errors.New(errors.CodeValidation,
			"unknown webhook type: "+string(req.Type),
		)
	}
}

func (p *IngestPipeline) normalizeDISStatusUpdate(ctx context.Context, correlationID string, req WebhookRequest) error {
	statusUpdate, err := p.normalizationEngine.NormalizeDISStatusUpdate(ctx, req.ProviderName, req.Payload)
	if err != nil {
		return err
	}

	// TODO (Orchestration): Hand statusUpdate to the OrchestratorEngine.
	// The orchestrator will: resolve the order, update canonical status,
	// trigger billing, comms, and WMS update — all as separate async tasks.
	// For now, log the normalized result as a development checkpoint.
	logger.Info(ctx, "DIS status update normalized",
		"correlation_id", correlationID,
		"provider_awb", statusUpdate.ProviderAWB,
		"provider_raw_status", statusUpdate.ProviderRawStatus,
		"canonical_status", statusUpdate.NewCanonicalStatus,
	)

	return nil
}

func (p *IngestPipeline) normalizeWMSOrderCreation(ctx context.Context, correlationID string, req WebhookRequest) error {
	normalizedOrder, err := p.normalizationEngine.NormalizeWMSOrder(ctx, req.ProviderName, req.Payload)
	if err != nil {
		return err
	}

	// TODO (Orchestration): Hand normalizedOrder to the OrchestratorEngine.
	// The orchestrator will persist the order, assign it a DIS, and dispatch.
	logger.Info(ctx, "WMS order creation normalized",
		"correlation_id", correlationID,
		"reference_code", normalizedOrder.ReferenceCode,
		"canonical_status", normalizedOrder.CanonicalStatus,
		"item_count", len(normalizedOrder.Items),
	)

	return nil
}

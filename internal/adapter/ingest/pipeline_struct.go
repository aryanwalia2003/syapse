// internal/adapter/ingest/pipeline_struct.go
package ingest

import (
	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/normalization"
)

// IngestPipeline is the central coordinator for the webhook ingest flow.
// It holds all dependencies injected at startup — it never constructs its
// own dependencies, making it fully testable in isolation.
//
// The pipeline is responsible for exactly three things, in order:
//  1. Persist the raw payload (audit trail, non-repudiation).
//  2. Dispatch to the correct normalizer via the NormalizationEngine.
//  3. Return a typed result for the HTTP handler to act on.
//
// It deliberately knows nothing about specific providers (Loginext, Uniware).
// That knowledge is sealed inside the adapters registered in the engine.
type IngestPipeline struct {
	normalizationEngine *normalization.NormalizationEngine
	webhookRepository   domain.WebhookRepository
}

// NewIngestPipeline constructs the pipeline with its required dependencies.
// Called once at startup in main.go, injected wherever needed.
func NewIngestPipeline(
	engine *normalization.NormalizationEngine,
	webhookRepo domain.WebhookRepository,
) *IngestPipeline {
	return &IngestPipeline{
		normalizationEngine: engine,
		webhookRepository:   webhookRepo,
	}
}

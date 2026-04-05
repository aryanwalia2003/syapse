// internal/adapter/ingest/webhook_handler.go
package ingest

import (
	"net/http"
	"strings"
	"time"

	"github.com/aryanwalia/synapse/internal/core/api"
	"github.com/aryanwalia/synapse/internal/core/errors"
)

// HandleDISWebhook returns an http.HandlerFunc for inbound DIS status webhooks.
// The provider name is read from the URL path segment (e.g. /webhook/dis/loginext).
func HandleDISWebhook(pipeline *IngestPipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerName := strings.ToUpper(r.PathValue("provider"))

		body, err := readAndLimitBody(r)
		if err != nil {
			api.WriteError(w, r, err)
			return
		}

		result := pipeline.Process(r.Context(), WebhookRequest{
			Type:              WebhookTypeDISStatusUpdate,
			ProviderName:      providerName,
			Payload:           body,
			Header:            r.Header,
			ReceivedAt:        time.Now().UTC(),
			VendorOrderIDPath: "orderNo", // Loginext sends the AWB as orderNo
		})

		if !result.Success {
			api.WriteError(w, r, errors.New(errors.CodeInternal, result.Message))
			return
		}

		api.WriteSuccess(w, r, result, result.CorrelationID)
	}
}

// HandleWMSWebhook returns an http.HandlerFunc for inbound WMS order creation webhooks.
func HandleWMSWebhook(pipeline *IngestPipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerName := strings.ToUpper(r.PathValue("provider"))

		body, err := readAndLimitBody(r)
		if err != nil {
			api.WriteError(w, r, err)
			return
		}

		result := pipeline.Process(r.Context(), WebhookRequest{
			Type:              WebhookTypeWMSOrderCreation,
			ProviderName:      providerName,
			Payload:           body,
			Header:            r.Header,
			ReceivedAt:        time.Now().UTC(),
			VendorOrderIDPath: "referenceNumber", // Uniware sends order ID as referenceNumber
		})

		if !result.Success {
			api.WriteError(w, r, errors.New(errors.CodeInternal, result.Message))
			return
		}

		api.WriteSuccess(w, r, result, result.CorrelationID)
	}
}

// readAndLimitBody reads the request body with a hard cap of 1MB.
// No webhook payload should ever exceed this; anything larger is likely abuse.
func readAndLimitBody(r *http.Request) ([]byte, error) {
	const maxBodyBytes = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)

	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for {
		n, readErr := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if readErr != nil {
			break
		}
	}

	if len(buf) == 0 {
		return nil, errors.New(errors.CodeValidation, "request body is empty")
	}
	return buf, nil
}

package ingest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/ingest/partition"
)

// capturingRepo captures the RawWebhook passed to SaveRaw for assertions.
type capturingRepo struct {
	domain.WebhookRepository
	captured *domain.RawWebhook
}

func (m *capturingRepo) SaveRaw(_ context.Context, webhook *domain.RawWebhook) error {
	m.captured = webhook
	return nil
}

func TestPersistRawWebhook_SetsPartitionFields(t *testing.T) {
	t.Run("WMS webhook sets VendorOrderID and PartitionIndex from payload", func(t *testing.T) {
		repo := &capturingRepo{}
		pipeline := NewIngestPipeline(nil, repo)
		handler := HandleWMSWebhook(pipeline)

		// Payload contains "referenceNumber" — the WMS vendor order ID
		payload := `{"referenceNumber": "ORD-ABC-42", "items": []}`
		req := httptest.NewRequest(http.MethodPost, "/webhook/wms/UNIWARE", bytes.NewBufferString(payload))
		req.SetPathValue("provider", "UNIWARE")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("Expected 202, got %d", rec.Code)
		}
		if repo.captured == nil {
			t.Fatal("Expected SaveRaw to be called")
		}
		if repo.captured.VendorOrderID != "ORD-ABC-42" {
			t.Errorf("Expected VendorOrderID=ORD-ABC-42, got %q", repo.captured.VendorOrderID)
		}
		expectedPartition := partition.HashPartition("ORD-ABC-42", 8)
		if repo.captured.PartitionIndex != expectedPartition {
			t.Errorf("Expected PartitionIndex=%d, got %d", expectedPartition, repo.captured.PartitionIndex)
		}
	})

	t.Run("DIS webhook sets VendorOrderID and PartitionIndex from payload", func(t *testing.T) {
		repo := &capturingRepo{}
		pipeline := NewIngestPipeline(nil, repo)
		handler := HandleDISWebhook(pipeline)

		// Loginext sends orderNo as the AWB/order identifier
		payload := `{"orderNo": "AWB-999", "status": "DELIVERED"}`
		req := httptest.NewRequest(http.MethodPost, "/webhook/dis/LOGINEXT", bytes.NewBufferString(payload))
		req.SetPathValue("provider", "LOGINEXT")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("Expected 202, got %d", rec.Code)
		}
		if repo.captured.VendorOrderID != "AWB-999" {
			t.Errorf("Expected VendorOrderID=AWB-999, got %q", repo.captured.VendorOrderID)
		}
		expectedPartition := partition.HashPartition("AWB-999", 8)
		if repo.captured.PartitionIndex != expectedPartition {
			t.Errorf("Expected PartitionIndex=%d, got %d", expectedPartition, repo.captured.PartitionIndex)
		}
	})

	t.Run("missing order ID sets unsorted partition (-1)", func(t *testing.T) {
		repo := &capturingRepo{}
		pipeline := NewIngestPipeline(nil, repo)
		handler := HandleWMSWebhook(pipeline)

		// Payload has no recognisable order ID field
		payload := `{"some_other_field": "value"}`
		req := httptest.NewRequest(http.MethodPost, "/webhook/wms/UNIWARE", bytes.NewBufferString(payload))
		req.SetPathValue("provider", "UNIWARE")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if repo.captured.VendorOrderID != "" {
			t.Errorf("Expected empty VendorOrderID, got %q", repo.captured.VendorOrderID)
		}
		if repo.captured.PartitionIndex != partition.UnsortedPartition() {
			t.Errorf("Expected PartitionIndex=%d, got %d", partition.UnsortedPartition(), repo.captured.PartitionIndex)
		}
	})
}

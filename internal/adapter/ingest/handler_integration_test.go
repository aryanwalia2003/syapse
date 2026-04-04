package ingest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func init() {
	logger.Init(slog.LevelDebug, false)
}

type mockHttpRepo struct {
	domain.WebhookRepository
	saveRawFail bool
	saveCalled  bool
}

func (m *mockHttpRepo) SaveRaw(ctx context.Context, webhook *domain.RawWebhook) error {
	m.saveCalled = true
	if m.saveRawFail {
		return context.DeadlineExceeded // simulating DB timeout
	}
	return nil
}

func TestHandlerIntegration(t *testing.T) {
	t.Run("returns 202 on successful DB write", func(t *testing.T) {
		repo := &mockHttpRepo{}
		pipeline := NewIngestPipeline(nil, repo)
		handler := HandleWMSWebhook(pipeline)

		req := httptest.NewRequest(http.MethodPost, "/webhook/wms/loginext", bytes.NewBufferString(`{"order_id": "123"}`))
		req.SetPathValue("provider", "LOGINEXT")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status 202, got %d", rec.Code)
		}
		if !repo.saveCalled {
			t.Errorf("Expected SaveRaw to be called")
		}
	})

	t.Run("returns error if DB write fails", func(t *testing.T) {
		repo := &mockHttpRepo{saveRawFail: true}
		pipeline := NewIngestPipeline(nil, repo)
		handler := HandleWMSWebhook(pipeline)

		req := httptest.NewRequest(http.MethodPost, "/webhook/wms/loginext", bytes.NewBufferString(`{"order_id": "123"}`))
		req.SetPathValue("provider", "LOGINEXT")
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code == http.StatusAccepted {
			t.Errorf("Expected error status, got %d", rec.Code)
		}
	})
}

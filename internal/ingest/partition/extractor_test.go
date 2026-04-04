package partition

import (
	"testing"
)

func TestExtractVendorOrderID(t *testing.T) {
	t.Run("extracts top level string key", func(t *testing.T) {
		payload := []byte(`{"order_id": "ORD-123", "status": "new"}`)
		id := ExtractVendorOrderID(payload, "order_id")
		if id != "ORD-123" {
			t.Errorf("Expected ORD-123, got %s", id)
		}
	})

	t.Run("extracts nested key", func(t *testing.T) {
		payload := []byte(`{"data": {"wms_id": "WMS-456"}}`)
		id := ExtractVendorOrderID(payload, "data.wms_id")
		if id != "WMS-456" {
			t.Errorf("Expected WMS-456, got %s", id)
		}
	})

	t.Run("returns empty on missing key", func(t *testing.T) {
		payload := []byte(`{"status": "new"}`)
		id := ExtractVendorOrderID(payload, "order_id")
		if id != "" {
			t.Errorf("Expected empty string, got %s", id)
		}
	})

	t.Run("returns empty on invalid json", func(t *testing.T) {
		payload := []byte(`{invalid`)
		id := ExtractVendorOrderID(payload, "order_id")
		if id != "" {
			t.Errorf("Expected empty string, got %s", id)
		}
	})
}

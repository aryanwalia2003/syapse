package normalization_test

import (
	"testing"

	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/normalization"
)

func baseNormalizedOrder() *domain.NormalizedOrder {
	return &domain.NormalizedOrder{
		ReferenceCode:      "REF-001",
		SourceProvider:     "UNIWARE",
		CanonicalStatus:    domain.CanonicalStatusPending,
		PackageWeightGrams: 500,
		Financials: domain.NormalizedFinancials{
			PaymentMode:      domain.PaymentModeCOD,
			TotalAmountPaise: 25000,
			Currency:         "INR",
		},
		Dimensions: domain.NormalizedDimensions{
			LengthCm: 10,
			WidthCm:  8,
			HeightCm: 5,
		},
		DeliveryAddress: domain.NormalizedAddress{
			ContactName:  "Riya Sharma",
			Phone:        "9999999999",
			AddressLine1: "42 Main St",
			AddressLine2: "Block B",
			City:         "Mumbai",
			State:        "Maharashtra",
			PinCode:      "400001",
		},
		Items: []domain.NormalizedOrderItem{
			{SKU: "SKU-A", Name: "Blue Shirt", Quantity: 2, PricePaise: 79900, WeightGrams: 200},
			{SKU: "SKU-B", Name: "Black Jeans", Quantity: 1, PricePaise: 149900, WeightGrams: 300},
		},
	}
}

func TestMapNormalizedOrderToComposite(t *testing.T) {
	t.Run("maps Order fields correctly", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		if result.Order.WMSOrderID != "REF-001" {
			t.Errorf("Expected WMSOrderID=REF-001, got %q", result.Order.WMSOrderID)
		}
		if result.Order.CanonicalStatus != string(domain.CanonicalStatusPending) {
			t.Errorf("Expected canonical status %q, got %q", domain.CanonicalStatusPending, result.Order.CanonicalStatus)
		}
		if result.Order.ID == "" {
			t.Error("Expected Order.ID to be set")
		}
	})

	t.Run("maps COD financials correctly", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		if result.Financials.PaymentMode != string(domain.PaymentModeCOD) {
			t.Errorf("Expected PaymentMode=COD, got %q", result.Financials.PaymentMode)
		}
		if result.Financials.CODAmountPaise != 25000 {
			t.Errorf("Expected CODAmountPaise=25000, got %d", result.Financials.CODAmountPaise)
		}
		if result.Financials.Currency != "INR" {
			t.Errorf("Expected Currency=INR, got %q", result.Financials.Currency)
		}
	})

	t.Run("maps PREPAID financials — COD amount is zero", func(t *testing.T) {
		n := baseNormalizedOrder()
		n.Financials.PaymentMode = domain.PaymentModePrepaid
		result := normalization.MapNormalizedOrderToComposite(n)

		if result.Financials.CODAmountPaise != 0 {
			t.Errorf("Expected CODAmountPaise=0 for prepaid, got %d", result.Financials.CODAmountPaise)
		}
	})

	t.Run("maps metrics correctly", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		if result.Metrics.WeightGrams != 500 {
			t.Errorf("Expected WeightGrams=500, got %d", result.Metrics.WeightGrams)
		}
		if result.Metrics.LengthCm != 10 || result.Metrics.WidthCm != 8 || result.Metrics.HeightCm != 5 {
			t.Errorf("Dimensions mismatch: got L=%d W=%d H=%d", result.Metrics.LengthCm, result.Metrics.WidthCm, result.Metrics.HeightCm)
		}
	})

	t.Run("maps recipient delivery address correctly", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		if result.Recipient.Name != "Riya Sharma" {
			t.Errorf("Expected Name=Riya Sharma, got %q", result.Recipient.Name)
		}
		if result.Recipient.Pincode != "400001" {
			t.Errorf("Expected Pincode=400001, got %q", result.Recipient.Pincode)
		}
		if result.Recipient.City != "Mumbai" {
			t.Errorf("Expected City=Mumbai, got %q", result.Recipient.City)
		}
		if result.Recipient.FullAddress == "" {
			t.Error("Expected FullAddress to be populated")
		}
	})

	t.Run("maps all items preserving SKU, quantity, and price", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		if len(result.Items) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(result.Items))
		}
		item := result.Items[0]
		if item.SKU != "SKU-A" {
			t.Errorf("Expected SKU=SKU-A, got %q", item.SKU)
		}
		if item.Quantity != 2 {
			t.Errorf("Expected Quantity=2, got %d", item.Quantity)
		}
		if item.PricePaise != 79900 {
			t.Errorf("Expected PricePaise=79900, got %d", item.PricePaise)
		}
		if item.ID == "" || item.OrderID == "" {
			t.Error("Expected item ID and OrderID to be set")
		}
	})

	t.Run("all sub-struct OrderIDs match the parent Order.ID", func(t *testing.T) {
		n := baseNormalizedOrder()
		result := normalization.MapNormalizedOrderToComposite(n)

		id := result.Order.ID
		if result.Financials.OrderID != id {
			t.Errorf("Financials.OrderID mismatch: got %q, want %q", result.Financials.OrderID, id)
		}
		if result.Metrics.OrderID != id {
			t.Errorf("Metrics.OrderID mismatch: got %q, want %q", result.Metrics.OrderID, id)
		}
		if result.Recipient.OrderID != id {
			t.Errorf("Recipient.OrderID mismatch: got %q, want %q", result.Recipient.OrderID, id)
		}
		for _, item := range result.Items {
			if item.OrderID != id {
				t.Errorf("Item.OrderID mismatch: got %q, want %q", item.OrderID, id)
			}
		}
	})
}

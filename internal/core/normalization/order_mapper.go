package normalization

import (
	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/pocketbase/pocketbase/tools/security"
)

// CompositeOrder holds all the entities that must be written atomically
// when a WMS order creation webhook is processed.
type CompositeOrder struct {
	Order      *domain.Order
	Financials *domain.OrderFinancials
	Metrics    *domain.OrderMetrics
	Recipient  *domain.OrderRecipient
	Items      []domain.OrderItem
}

// MapNormalizedOrderToComposite converts a validated NormalizedOrder into the
// set of domain entities required by OrderRepository.CreateCompositeOrder.
// It generates a single orderID that is stamped on every sub-entity so that
// all foreign keys are consistent.
func MapNormalizedOrderToComposite(n *domain.NormalizedOrder) *CompositeOrder {
	orderID := "ORD-" + security.RandomString(10)

	return &CompositeOrder{
		Order:      mapOrder(n, orderID),
		Financials: mapFinancials(n, orderID),
		Metrics:    mapMetrics(n, orderID),
		Recipient:  mapRecipient(n, orderID),
		Items:      mapItems(n, orderID),
	}
}

func mapOrder(n *domain.NormalizedOrder, orderID string) *domain.Order {
	return &domain.Order{
		ID:              orderID,
		BrandID:         "DEFAULT_BRAND",
		WarehouseID:     "DEFAULT_WAREHOUSE",
		WMSOrderID:      n.ReferenceCode,
		CanonicalStatus: string(n.CanonicalStatus),
	}
}

func mapFinancials(n *domain.NormalizedOrder, orderID string) *domain.OrderFinancials {
	var codAmount int64
	if n.Financials.PaymentMode == domain.PaymentModeCOD {
		codAmount = n.Financials.TotalAmountPaise
	}

	return &domain.OrderFinancials{
		OrderID:        orderID,
		PaymentMode:    string(n.Financials.PaymentMode),
		CODAmountPaise: codAmount,
		Currency:       n.Financials.Currency,
	}
}

func mapMetrics(n *domain.NormalizedOrder, orderID string) *domain.OrderMetrics {
	return &domain.OrderMetrics{
		OrderID:     orderID,
		WeightGrams: n.PackageWeightGrams,
		LengthCm:    n.Dimensions.LengthCm,
		WidthCm:     n.Dimensions.WidthCm,
		HeightCm:    n.Dimensions.HeightCm,
	}
}

func mapRecipient(n *domain.NormalizedOrder, orderID string) *domain.OrderRecipient {
	addr := n.DeliveryAddress
	return &domain.OrderRecipient{
		OrderID:     orderID,
		Name:        addr.ContactName,
		Phone:       addr.Phone,
		City:        addr.City,
		State:       addr.State,
		Pincode:     addr.PinCode,
		FullAddress: addr.AddressLine1 + " " + addr.AddressLine2,
	}
}

func mapItems(n *domain.NormalizedOrder, orderID string) []domain.OrderItem {
	items := make([]domain.OrderItem, 0, len(n.Items))
	for _, item := range n.Items {
		items = append(items, domain.OrderItem{
			ID:         "ITEM-" + security.RandomString(10),
			OrderID:    orderID,
			SKU:        item.SKU,
			Name:       item.Name,
			Quantity:   item.Quantity,
			PricePaise: item.PricePaise,
		})
	}
	return items
}

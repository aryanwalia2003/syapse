package order

import (
	"context"

	"github.com/aryanwalia/synapse/internal/core/db"
	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/pocketbase/pocketbase/tools/security"
)

func (repo *SQLOrderRepository) CreateCompositeOrder(
	ctx context.Context,
	order *domain.Order,
	financials *domain.OrderFinancials,
	metrics *domain.OrderMetrics,
	recipient *domain.OrderRecipient,
	items []domain.OrderItem,
) error {
	return repo.db.RunInTransaction(ctx, func(tx db.Transaction) error {

		if order.ID == "" {
			order.ID = security.RandomString(15)
		}

		orderSQL := `
			INSERT INTO orders (id, brand_id, warehouse_id, wms_order_id, canonical_status)
			VALUES ({:id}, {:brand_id}, {:warehouse_id}, {:wms_order_id}, {:canonical_status})
		`
		err := tx.Execute(ctx, orderSQL, map[string]any{
			"id":               order.ID,
			"brand_id":         order.BrandID,
			"warehouse_id":     order.WarehouseID,
			"wms_order_id":     order.WMSOrderID,
			"canonical_status": order.CanonicalStatus,
		})
		if err != nil {
			return err
		}

		if financials != nil {
			finSQL := `
				INSERT INTO order_financials (id, order_id, payment_mode, total_amount_paise)
				VALUES ({:id}, {:order_id}, {:payment_mode}, {:total_amount_paise})
			`
			err = tx.Execute(ctx, finSQL, map[string]any{
				"id":                 security.RandomString(15),
				"order_id":           order.ID,
				"payment_mode":       financials.PaymentMode,
				"total_amount_paise": financials.CODAmountPaise,
			})
			if err != nil {
				return err
			}
		}

		if recipient != nil {
			recSQL := `
				INSERT INTO order_recipients (id, order_id, name, phone, pincode)
				VALUES ({:id}, {:order_id}, {:name}, {:phone}, {:pincode})
			`
			err = tx.Execute(ctx, recSQL, map[string]any{
				"id":       security.RandomString(15),
				"order_id": order.ID,
				"name":     recipient.Name,
				"phone":    recipient.Phone,
				"pincode":  recipient.Pincode,
			})
			if err != nil {
				return err
			}
		}

		if len(items) > 0 {
			itemSQL := `
				INSERT INTO order_items (id, order_id, sku, name, quantity, price_paise)
				VALUES ({:id}, {:order_id}, {:sku}, {:name}, {:quantity}, {:price_paise})
			`
			for _, item := range items {
				err = tx.Execute(ctx, itemSQL, map[string]any{
					"id":          security.RandomString(15),
					"order_id":    order.ID,
					"sku":         item.SKU,
					"name":        item.Name,
					"quantity":    item.Quantity,
					"price_paise": item.PricePaise,
				})
				if err != nil {
					return err
				}
			}
		}

		if metrics != nil {
			metSQL := `
				INSERT INTO order_metrics (id, order_id, dead_weight_grams, length_cm, width_cm, height_cm)
				VALUES ({:id}, {:order_id}, {:dead_weight_grams}, {:length_cm}, {:width_cm}, {:height_cm})
			`
			err = tx.Execute(ctx, metSQL, map[string]any{
				"id":                security.RandomString(15),
				"order_id":          order.ID,
				"dead_weight_grams": metrics.WeightGrams,
				"length_cm":         metrics.LengthCm,
				"width_cm":          metrics.WidthCm,
				"height_cm":         metrics.HeightCm,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

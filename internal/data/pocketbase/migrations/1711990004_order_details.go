package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		orders, _ := app.FindCollectionByNameOrId("orders")

		// 8. Order Items
		items := core.NewBaseCollection("order_items")
		items.Fields.Add(
			&core.RelationField{Name: "order_id", Required: true, CollectionId: orders.Id, MaxSelect: 1},
			&core.TextField{Name: "sku", Required: true},
			&core.TextField{Name: "name"},
			&core.NumberField{Name: "quantity"},
			&core.NumberField{Name: "price_paise"},
		)
		if err := app.Save(items); err != nil {
			return err
		}

		// 9. Order Financials
		financials := core.NewBaseCollection("order_financials")
		financials.Fields.Add(
			&core.RelationField{Name: "order_id", Required: true, CollectionId: orders.Id, MaxSelect: 1, Id: "order_fin_id"},
			&core.NumberField{Name: "total_amount_paise"},
			&core.NumberField{Name: "discount_amount_paise"},
			&core.NumberField{Name: "shipping_amount_paise"},
			&core.SelectField{Name: "payment_mode", Values: []string{"PREPAID", "COD"}},
		)
		if err := app.Save(financials); err != nil {
			return err
		}

		// 10. Order Metrics
		metrics := core.NewBaseCollection("order_metrics")
		metrics.Fields.Add(
			&core.RelationField{Name: "order_id", Required: true, CollectionId: orders.Id, MaxSelect: 1},
			&core.NumberField{Name: "dead_weight_grams"},
			&core.NumberField{Name: "length_cm"},
			&core.NumberField{Name: "width_cm"},
			&core.NumberField{Name: "height_cm"},
		)
		return app.Save(metrics)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("order_metrics")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("order_financials")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("order_items")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		brands, _ := app.FindCollectionByNameOrId("brands")
		warehouses, _ := app.FindCollectionByNameOrId("warehouses")

		// 4. Orders
		orders := core.NewBaseCollection("orders")
		orders.Fields.Add(
			&core.RelationField{Name: "brand_id", Required: true, CollectionId: brands.Id, MaxSelect: 1},
			&core.RelationField{Name: "warehouse_id", Required: true, CollectionId: warehouses.Id, MaxSelect: 1},
			&core.TextField{Name: "wms_order_id", Required: true},
			&core.TextField{Name: "dis_awb"},
			&core.TextField{Name: "canonical_status", Required: true},
		)
		if err := app.Save(orders); err != nil {
			return err
		}

		// 5. Order Recipients
		recipient := core.NewBaseCollection("order_recipients")
		recipient.Fields.Add(
			&core.RelationField{Name: "order_id", Required: true, CollectionId: orders.Id, MaxSelect: 1},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "pincode"},
			&core.TextField{Name: "full_address"},
		)
		return app.Save(recipient)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("order_recipients")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("orders")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

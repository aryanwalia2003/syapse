package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		// 1. Brands Collection
		brands := core.NewBaseCollection("brands")
		brands.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.EmailField{Name: "contact_email"},
			&core.BoolField{Name: "is_active"},
		)
		if err := app.Save(brands); err != nil {
			return err
		}

		// 2. Warehouses Collection
		warehouses := core.NewBaseCollection("warehouses")
		warehouses.Fields.Add(
			&core.RelationField{
				Name:         "brand_id",
				Required:     true,
				CollectionId: brands.Id,
				MaxSelect:    1,
			},
			&core.TextField{Name: "facility_code", Required: true},
			&core.TextField{Name: "name", Required: true},
			&core.TextField{Name: "pincode"},
			&core.TextField{Name: "city"},
			&core.TextField{Name: "state"},
			&core.BoolField{Name: "is_active"},
		)
		return app.Save(warehouses)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("warehouses")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("brands")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

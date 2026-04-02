package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		brands, _ := app.FindCollectionByNameOrId("brands")
		wmsProviders, _ := app.FindCollectionByNameOrId("wms_providers")
		disProviders, _ := app.FindCollectionByNameOrId("dis_providers")

		// 6. WMS Integrations
		wmsInteg := core.NewBaseCollection("wms_integrations")
		wmsInteg.Fields.Add(
			&core.RelationField{Name: "brand_id", Required: true, CollectionId: brands.Id, MaxSelect: 1},
			&core.RelationField{Name: "wms_provider_id", Required: true, CollectionId: wmsProviders.Id, MaxSelect: 1},
			&core.TextField{Name: "api_credentials"}, // Encrypted JSON
			&core.BoolField{Name: "is_active"},
		)
		if err := app.Save(wmsInteg); err != nil {
			return err
		}

		// 7. DIS Integrations
		disInteg := core.NewBaseCollection("dis_integrations")
		disInteg.Fields.Add(
			&core.RelationField{Name: "brand_id", Required: true, CollectionId: brands.Id, MaxSelect: 1},
			&core.RelationField{Name: "dis_provider_id", Required: true, CollectionId: disProviders.Id, MaxSelect: 1},
			&core.TextField{Name: "api_credentials"}, // Encrypted JSON
			&core.TextField{Name: "webhook_secret"},
			&core.BoolField{Name: "is_active"},
		)
		return app.Save(disInteg)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("dis_integrations")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("wms_integrations")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

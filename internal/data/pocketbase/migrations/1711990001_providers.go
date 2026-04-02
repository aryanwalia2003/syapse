package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		// 3. Providers
		wmsProviders := core.NewBaseCollection("wms_providers")
		wmsProviders.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.BoolField{Name: "is_supported"},
		)
		if err := app.Save(wmsProviders); err != nil {
			return err
		}

		disProviders := core.NewBaseCollection("dis_providers")
		disProviders.Fields.Add(
			&core.TextField{Name: "name", Required: true},
			&core.BoolField{Name: "is_supported"},
		)
		return app.Save(disProviders)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("dis_providers")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("wms_providers")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

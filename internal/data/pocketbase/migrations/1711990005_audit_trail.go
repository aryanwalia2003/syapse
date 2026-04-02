package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		orders, _ := app.FindCollectionByNameOrId("orders")

		// 11. Order State Logs
		logs := core.NewBaseCollection("order_state_logs")
		logs.Fields.Add(
			&core.RelationField{Name: "order_id", Required: true, CollectionId: orders.Id, MaxSelect: 1},
			&core.TextField{Name: "correlation_id"},
			&core.TextField{Name: "prev_status"},
			&core.TextField{Name: "new_status"},
			&core.TextField{Name: "triggered_by"},
		)
		if err := app.Save(logs); err != nil {
			return err
		}

		// 12. Raw Webhook Payloads (Audit Trail)
		raw := core.NewBaseCollection("raw_webhook_payloads")
		raw.Fields.Add(
			&core.TextField{Name: "correlation_id", Required: true},
			&core.TextField{Name: "source", Required: true},
			&core.JSONField{Name: "payload"},
			&core.JSONField{Name: "headers"},
		)
		return app.Save(raw)
	}, func(app core.App) error {
		c, _ := app.FindCollectionByNameOrId("raw_webhook_payloads")
		if c != nil {
			app.Delete(c)
		}
		c, _ = app.FindCollectionByNameOrId("order_state_logs")
		if c != nil {
			app.Delete(c)
		}
		return nil
	})
}

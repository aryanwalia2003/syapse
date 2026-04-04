package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		// Add outbox pattern fields to raw_webhook_payloads
		rawCollection, err := app.FindCollectionByNameOrId("raw_webhook_payloads")
		if err != nil {
			return err
		}

		rawCollection.Fields.Add(
			&core.TextField{Name: "status"},            // PENDING | PROCESSING | DONE | FAILED
			&core.TextField{Name: "vendor_order_id"},   // extracted from payload for partitioning
			&core.TextField{Name: "vendor_webhook_id"}, // idempotency key
			&core.NumberField{Name: "retry_count"},     // 0..maxRetries
			&core.BoolField{Name: "is_dlq"},            // true after max retries exceeded
			&core.NumberField{Name: "partition_index"}, // hash(vendor_order_id) % N, or -1
		)

		return app.Save(rawCollection)
	}, func(app core.App) error {
		// Down: remove the columns
		rawCollection, err := app.FindCollectionByNameOrId("raw_webhook_payloads")
		if err != nil {
			return err
		}

		for _, name := range []string{"status", "vendor_order_id", "vendor_webhook_id", "retry_count", "is_dlq", "partition_index"} {
			rawCollection.Fields.RemoveByName(name)
		}

		return app.Save(rawCollection)
	})
}

package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BilledResource.Quantity
		_, err = db.Exec(`
			ALTER TABLE billed_resources
			ALTER COLUMN quantity SET DATA TYPE real;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BilledResource.Quantity
		_, err = db.Exec(`
			ALTER TABLE billed_resources
			ALTER COLUMN quantity SET DATA TYPE bigint;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

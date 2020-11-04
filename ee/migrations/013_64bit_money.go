package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ALTER COLUMN base_price_cents SET DATA TYPE bigint,
			ALTER COLUMN seat_price_cents SET DATA TYPE bigint,
			ALTER COLUMN read_overage_price_cents SET DATA TYPE bigint,
			ALTER COLUMN write_overage_price_cents SET DATA TYPE bigint,
			ALTER COLUMN scan_overage_price_cents SET DATA TYPE bigint;

			ALTER TABLE billed_resources
			ALTER COLUMN total_price_cents SET DATA TYPE bigint,
			ALTER COLUMN quantity SET DATA TYPE double precision;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ALTER COLUMN base_price_cents SET DATA TYPE integer,
			ALTER COLUMN seat_price_cents SET DATA TYPE integer,
			ALTER COLUMN read_overage_price_cents SET DATA TYPE integer,
			ALTER COLUMN write_overage_price_cents SET DATA TYPE integer,
			ALTER COLUMN scan_overage_price_cents SET DATA TYPE integer;

			ALTER TABLE billed_resources
			ALTER COLUMN total_price_cents SET DATA TYPE integer,
			ALTER COLUMN quantity SET DATA TYPE real;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

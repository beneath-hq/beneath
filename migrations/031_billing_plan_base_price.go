package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.BasePriceCents
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD base_price_cents integer NOT NULL default 0;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.BasePriceCents
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP base_price_cents;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

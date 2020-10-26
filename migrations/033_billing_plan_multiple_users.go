package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.MultipleUsers
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			RENAME COLUMN personal TO multiple_users;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.MultipleUsers
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			RENAME COLUMN multiple_users TO personal;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

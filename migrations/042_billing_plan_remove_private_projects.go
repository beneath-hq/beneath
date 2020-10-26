package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.PrivateProjects
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP private_projects;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.PrivateProjects
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD private_projects boolean NOT NULL default true;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

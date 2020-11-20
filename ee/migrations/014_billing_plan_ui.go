package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.Name
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD name text NOT NULL default 'TODO';
		`)
		if err != nil {
			return err
		}

		// BillingPlan.UIRank
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD ui_rank integer;
			CREATE INDEX ON billing_plans (ui_rank) WHERE ui_rank IS NOT NULL;
		`)
		if err != nil {
			return err
		}

		// BillingPlan.AvailableInUI
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP available_in_ui;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.AvailableInUI
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD available_in_ui boolean NOT NULL default false;
		`)
		if err != nil {
			return err
		}

		// BillingPlan.UIRank
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP ui_rank;
		`)
		if err != nil {
			return err
		}

		// BillingPlan.Name
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP name;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

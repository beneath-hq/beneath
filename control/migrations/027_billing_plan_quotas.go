package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.ReadQuotaCap, BillingPlan.WriteQuotaCap
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD read_quota bigint NOT NULL default 0,
			ADD write_quota bigint NOT NULL default 0;
		`)
		if err != nil {
			return err
		}

		// migrate data
		_, err = db.Exec(`
			update billing_plans
			set read_quota = base_read_quota,
					write_quota = base_write_quota;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.ReadQuotaCap, BillingPlan.WriteQuotaCap
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP read_quota,
			DROP write_quota;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.ReadQuotaCap, BillingPlan.WriteQuotaCap
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD prepaid_read_quota bigint,
			ADD prepaid_write_quota bigint;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.ReadQuotaCap, BillingPlan.WriteQuotaCap
		_, err = db.Exec(`
			ALTER TABLE organizations
			DROP prepaid_read_quota,
			DROP prepaid_write_quota;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

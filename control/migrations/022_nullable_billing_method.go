package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingInfo.BillingMethodID
		_, err = db.Exec(`
			ALTER TABLE billing_infos
			ALTER COLUMN billing_method_id DROP NOT NULL;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Done
		return nil
	})
}

package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE billing_infos
			ADD next_billing_time timestamptz NOT NULL DEFAULT now(),
			ADD last_invoice_time timestamptz NOT NULL DEFAULT now()
			;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE billing_infos
			DROP next_billing_time,
			DROP last_invoice_time
			;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

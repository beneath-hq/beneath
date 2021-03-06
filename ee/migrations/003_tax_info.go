package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingInfo tax info
		_, err = db.Exec(`
			ALTER TABLE billing_infos
			ADD country text,
			ADD region text,
			ADD company_name text,
			ADD tax_number text;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingInfo tax info
		_, err = db.Exec(`
			ALTER TABLE billing_infos 
			DROP country,
			DROP region,
			DROP company_name,
			DROP tax_number;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

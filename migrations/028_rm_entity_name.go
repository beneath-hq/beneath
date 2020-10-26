package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BilledResource.EntityName
		_, err = db.Exec(`
			ALTER TABLE billed_resources
			DROP entity_name;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BilledResource.EntityName
		_, err = db.Exec(`
			ALTER TABLE billed_resources
			ADD entity_name text NOT NULL default 'TODO';
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

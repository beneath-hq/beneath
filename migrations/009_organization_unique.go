package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Organization.Name UNIQUE
		_, err = db.Exec(`
			CREATE UNIQUE INDEX organizations_name_key ON public.organizations USING btree (lower(name));
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Organization.Name UNIQUE
		_, err = db.Exec(`
			DROP INDEX organizations_name_key;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

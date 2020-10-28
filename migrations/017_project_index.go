package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Project
		// drop (name) index
		_, err = db.Exec(`
			DROP INDEX projects_name_key;
		`)
		if err != nil {
			return err
		}

		// add (Organization, name) index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX projects_organization_id_name_key ON public.projects USING btree (organization_id, (lower(name)));
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Project
		// drop (Organization, name) index
		_, err = db.Exec(`
			DROP INDEX projects_organization_id_name_key;
		`)
		if err != nil {
			return err
		}

		// add (name) index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX projects_name_key ON public.projects USING btree ((lower(name)));
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

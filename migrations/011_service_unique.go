package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Service.Name NOT NULL
		_, err = db.Exec(`
			ALTER TABLE services ALTER COLUMN name SET NOT NULL;
		`)
		if err != nil {
			return err
		}

		// (Organization, name) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX services_organization_id_name_key ON public.services USING btree (organization_id, (lower(name)));
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// (Organization, name) unique index
		_, err = db.Exec(`
			DROP INDEX IF EXISTS services_organization_id_name_key;
		`)
		if err != nil {
			return err
		}

		// Service.Name NOT NULL
		_, err = db.Exec(`
			ALTER TABLE services ALTER COLUMN name DROP NOT NULL;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

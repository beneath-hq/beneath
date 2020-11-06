package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Extensions
		_, err = db.Exec(`
			CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		`)
		if err != nil {
			return err
		}

		// Organization
		_, err = db.Exec(`
			CREATE TABLE organizations (
				organization_id uuid DEFAULT uuid_generate_v4(),
				name text NOT NULL,
				created_on timestamp with time zone DEFAULT now(),
				updated_on timestamp with time zone DEFAULT now(),
				PRIMARY KEY (organization_id)
			);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Organization
		_, err = db.Exec(`
			DROP TABLE IF EXISTS organizations CASCADE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

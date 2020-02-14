package migrations

import (
	"github.com/beneath-core/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
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
		err = db.Model(&entity.Organization{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Extensions
		_, err = db.Exec(`
			DROP EXTENSION "uuid-ossp";
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

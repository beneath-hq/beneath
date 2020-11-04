package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Organization.Personal
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD personal bool NOT NULL default FALSE;
		`)
		if err != nil {
			return err
		}

		// User.OrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN main_organization_id TO organization_id;
		`)
		if err != nil {
			return err
		}

		// Project.Locked
		_, err = db.Exec(`
			ALTER TABLE projects ADD locked bool NOT NULL default FALSE;
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(db migrations.DB) (err error) {
		// Project.Locked
		_, err = db.Exec(`
			ALTER TABLE projects DROP locked;
		`)
		if err != nil {
			return err
		}

		// User.OrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN organization_id TO main_organization_id;
		`)
		if err != nil {
			return err
		}

		// Organization.Personal
		_, err = db.Exec(`
			ALTER TABLE organizations DROP personal;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}

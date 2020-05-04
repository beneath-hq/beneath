package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// PermissionsUsersOrganizations.Create
		_, err = db.Exec(`
			ALTER TABLE permissions_users_organizations
			ADD "create" boolean NOT NULL DEFAULT false;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// PermissionsUsersOrganizations.Create
		_, err = db.Exec(`
			ALTER TABLE permissions_users_organizations 
			DROP "create";
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

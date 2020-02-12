package migrations

import (
	"github.com/beneath-core/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// User
		err = db.Model(&entity.User{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Username unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX users_username_key ON public.users USING btree ((lower(username)));
		`)
		if err != nil {
			return err
		}

		// User email unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX users_email_key ON public.users USING btree ((lower(email)));
		`)
		if err != nil {
			return err
		}

		// PermissionsUsersProjects
		err = db.Model(&entity.PermissionsUsersProjects{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// PermissionsUsersOrganizations
		err = db.Model(&entity.PermissionsUsersOrganizations{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// PermissionsUsersProjects
		err = db.Model(&entity.PermissionsUsersProjects{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// PermissionsUsersOrganizations
		err = db.Model(&entity.PermissionsUsersOrganizations{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// User
		err = db.Model(&entity.User{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

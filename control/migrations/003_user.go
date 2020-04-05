package migrations

import (
	"gitlab.com/beneath-org/beneath/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// User
		_, err = db.Exec(`
			CREATE TABLE users (
				user_id uuid DEFAULT uuid_generate_v4(),
				username text NOT NULL,
				email text NOT NULL,
				name text NOT NULL,
				bio text,
				photo_url text,
				google_id text UNIQUE,
				github_id text UNIQUE,
				created_on timestamptz DEFAULT now(),
				updated_on timestamptz DEFAULT now(),
				main_organization_id uuid NOT NULL,
				read_quota bigint,
				write_quota bigint,
				PRIMARY KEY (user_id),
				FOREIGN KEY (main_organization_id) REFERENCES organizations (organization_id) ON DELETE restrict
			);
		`)
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
		_, err = db.Exec(`
			CREATE TABLE permissions_users_projects (
				user_id uuid,
				project_id uuid,
				view boolean NOT NULL,
				"create" boolean NOT NULL,
				admin boolean NOT NULL,
				PRIMARY KEY (user_id, project_id),
				FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
				FOREIGN KEY (project_id) REFERENCES projects (project_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		// PermissionsUsersOrganizations
		_, err = db.Exec(`
			CREATE TABLE permissions_users_organizations (
				user_id uuid,
				organization_id uuid,
				view boolean NOT NULL,
				admin boolean NOT NULL,
				PRIMARY KEY (user_id, organization_id),
				FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
				FOREIGN KEY (organization_id) REFERENCES organizations (organization_id) ON DELETE CASCADE
			);
		`)
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

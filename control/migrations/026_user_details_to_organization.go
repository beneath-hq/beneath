package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Organizations
		_, err = db.Exec(`
			ALTER TABLE organizations
			DROP personal,
			ADD user_id UUID,
			ADD FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE RESTRICT,
			ADD display_name TEXT,
			ADD description TEXT,
			ADD photo_url TEXT;
		`)
		if err != nil {
			return err
		}

		// migrate data
		_, err = db.Exec(`
			update organizations o
			set
				user_id = u.user_id,
				display_name = u.name,
				description = u.bio,
				photo_url = u.photo_url
			from users u 
			where o.organization_id = u.personal_organization_id
			;
		`)
		if err != nil {
			return err
		}

		// Organizations.DisplayName not null
		_, err = db.Exec(`
			ALTER TABLE organizations ALTER COLUMN display_name SET NOT NULL
		`)
		if err != nil {
			return err
		}

		// Users
		_, err = db.Exec(`
			ALTER TABLE users
			DROP username,
			DROP name,
			DROP bio,
			DROP photo_url,
			DROP personal_organization_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Organizations
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD personal bool NOT NULL default FALSE,
			DROP user_id,
			DROP display_name,
			DROP description,
			DROP photo_url;
		`)
		if err != nil {
			return err
		}

		// Users
		_, err = db.Exec(`
			ALTER TABLE users
			ADD username TEXT,
			ADD name TEXT,
			ADD bio TEXT,
			ADD photo_url TEXT,
			ADD personal_organization_id uuid REFERENCES organizations (organization_id) ON DELETE restrict;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

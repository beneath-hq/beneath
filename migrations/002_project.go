package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Project
		_, err = db.Exec(`
			CREATE TABLE projects (
				project_id uuid DEFAULT uuid_generate_v4(),
				name text NOT NULL,
				display_name text NOT NULL,
				site text,
				description text,
				photo_url text,
				public boolean NOT NULL DEFAULT true,
				organization_id uuid NOT NULL,
				created_on timestamp with time zone DEFAULT now(),
				updated_on timestamp with time zone DEFAULT now(),
				PRIMARY KEY (project_id),
				FOREIGN KEY (organization_id) REFERENCES organizations (organization_id) ON DELETE restrict
			);
		`)
		if err != nil {
			return err
		}

		// Project name index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX projects_name_key ON public.projects USING btree ((lower(name)));
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			DROP TABLE IF EXISTS projects CASCADE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

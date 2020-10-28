package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE services
			ADD description TEXT,
			ADD source_url TEXT,
			ADD project_id uuid REFERENCES projects (project_id) ON DELETE restrict;

			CREATE UNIQUE INDEX ON public.services USING btree (project_id, (lower(name)));

			UPDATE services s
			SET project_id = pp.project_id
			FROM (
				SELECT DISTINCT ON (p.organization_id) p.organization_id, p.project_id
				FROM projects p
			) pp
			WHERE pp.organization_id = s.organization_id;

			ALTER TABLE services
			DROP kind,
			DROP organization_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE services
			ADD kind text,
			ADD organization_id uuid REFERENCES organizations (organization_id) ON DELETE restrict;

			CREATE UNIQUE INDEX services_organization_id_name_key ON public.services USING btree (organization_id, (lower(name)));

			ALTER TABLE services
			DROP description,
			DROP source_url,
			DROP project_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

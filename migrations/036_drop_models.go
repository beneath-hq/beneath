package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams DROP source_model_id;
			DROP TABLE streams_into_models;
			DROP TABLE models;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			CREATE TABLE models (
				model_id uuid DEFAULT uuid_generate_v4(),
				name text NOT NULL,
				description text,
				source_url text,
				kind text NOT NULL,
				created_on timestamptz DEFAULT now(),
				updated_on timestamptz DEFAULT now(),
				project_id uuid NOT NULL,
				service_id uuid NOT NULL,
				PRIMARY KEY (model_id),
				FOREIGN KEY (project_id) REFERENCES projects (project_id) ON DELETE RESTRICT,
				FOREIGN KEY (service_id) REFERENCES services (service_id) ON DELETE RESTRICT
			);
		`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
			CREATE TABLE streams_into_models (
				stream_id uuid,
				model_id uuid,
				PRIMARY KEY (stream_id, model_id),
				FOREIGN KEY (stream_id) REFERENCES streams (stream_id) ON DELETE CASCADE,
				FOREIGN KEY (model_id) REFERENCES models (model_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
			ALTER TABLE streams
			ADD source_model_id UUID,
			ADD FOREIGN KEY (source_model_id) REFERENCES models (model_id) ON DELETE RESTRICT;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

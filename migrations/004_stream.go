package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream
		_, err = db.Exec(`
			CREATE TABLE streams(
				stream_id             UUID DEFAULT uuid_generate_v4(),
				name                  TEXT NOT NULL,
				description           TEXT,
				schema                TEXT NOT NULL,
				avro_schema           JSON NOT NULL,
				canonical_avro_schema JSON NOT NULL,
				bigquery_schema       JSON NOT NULL,
				key_fields            JSONB NOT NULL, 
				external              BOOLEAN NOT NULL,
				batch                 BOOLEAN NOT NULL,
				manual                BOOLEAN NOT NULL,
				project_id            UUID NOT NULL,
				created_on            TIMESTAMPTZ DEFAULT Now(),
				updated_on            TIMESTAMPTZ DEFAULT Now(),
				PRIMARY KEY (stream_id),
				FOREIGN KEY (project_id) REFERENCES projects (project_id) ON DELETE RESTRICT
			);
		`)
		if err != nil {
			return err
		}

		// (Project, name) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX streams_project_id_name_key ON public.streams USING btree (project_id, (lower(name)));
		`)
		if err != nil {
			return err
		}

		// StreamInstance
		_, err = db.Exec(`
			CREATE TABLE stream_instances(
				stream_instance_id uuid DEFAULT uuid_generate_v4(),
				stream_id uuid NOT NULL,
				created_on timestamptz DEFAULT now(),
				updated_on timestamptz DEFAULT now(),
				committed_on timestamptz,
				PRIMARY KEY (stream_instance_id),
				FOREIGN KEY (stream_id) REFERENCES streams (stream_id) ON DELETE RESTRICT
			);
		`)
		if err != nil {
			return err
		}

		// Stream.CurrentStreamInstanceID
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD current_stream_instance_id UUID,
			ADD FOREIGN KEY (current_stream_instance_id) REFERENCES stream_instances (stream_instance_id);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			DROP TABLE IF EXISTS stream_instances CASCADE;
			DROP TABLE IF EXISTS streams CASCADE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
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
				deleted_on            TIMESTAMPTZ,
				PRIMARY KEY (stream_id),
				FOREIGN KEY (project_id) REFERENCES projects (project_id) ON DELETE RESTRICT
			)
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
		err = db.Model(&entity.StreamInstance{}).CreateTable(defaultCreateOptions)
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
		// StreamInstance
		err = db.Model(&entity.StreamInstance{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Stream
		err = db.Model(&entity.Stream{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

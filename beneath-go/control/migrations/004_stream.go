package migrations

import (
	"github.com/beneath-core/beneath-go/control/model"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream
		_, err = db.Exec(`
			CREATE TABLE streams(
				stream_id    UUID,
				name         TEXT NOT NULL,
				description  TEXT,
				schema       TEXT NOT NULL,
				avro_schema  JSON,
				external     BOOLEAN NOT NULL,
				batch        BOOLEAN NOT NULL,
				manual       BOOLEAN NOT NULL,
				project_id   UUID NOT NULL,
				created_on   TIMESTAMPTZ DEFAULT Now(),
				updated_on   TIMESTAMPTZ DEFAULT Now(),
				PRIMARY KEY (stream_id),
				UNIQUE (project_id, name),
				FOREIGN KEY (project_id) REFERENCES projects (project_id) ON DELETE RESTRICT
			)
		`)
		if err != nil {
			return err
		}

		// StreamInstance
		err = db.Model(&model.StreamInstance{}).CreateTable(defaultCreateOptions)
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
		err = db.Model(&model.StreamInstance{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Stream
		err = db.Model(&model.Stream{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

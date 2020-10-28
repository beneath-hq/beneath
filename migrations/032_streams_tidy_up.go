package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream
		_, err = db.Exec(`
			ALTER TABLE streams
			DROP external,
			DROP batch,
			ADD schema_kind text not null default '',
			ADD schema_md5 bytea not null default '',
			ADD instances_deleted_count integer NOT NULL DEFAULT 0,
			ADD instances_made_final_count integer NOT NULL DEFAULT 0;
			
			ALTER TABLE streams RENAME COLUMN current_stream_instance_id TO primary_stream_instance_id;
			ALTER TABLE streams RENAME COLUMN manual TO enable_manual_writes;
			ALTER TABLE streams RENAME COLUMN instances_committed_count TO instances_made_primary_count;

			ALTER TABLE streams
			DROP CONSTRAINT streams_current_stream_instance_id_fkey,
			ADD CONSTRAINT streams_primary_stream_instance_id_fkey
				FOREIGN KEY (primary_stream_instance_id)
				REFERENCES stream_instances(stream_instance_id)
				ON DELETE SET NULL;
		`)
		if err != nil {
			return err
		}

		// StreamInstance
		_, err = db.Exec(`
			ALTER TABLE stream_instances ADD made_final_on timestamptz;
			ALTER TABLE stream_instances RENAME COLUMN committed_on TO made_primary_on;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Stream
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD external BOOLEAN,
			ADD batch BOOLEAN,
			DROP schema_kind,
			DROP schema_md5,
			DROP instances_deleted_count,
			DROP instances_made_final_count;

			ALTER TABLE streams RENAME COLUMN primary_stream_instance_id TO current_stream_instance_id;
			ALTER TABLE streams RENAME COLUMN enable_manual_writes TO manual;
			ALTER TABLE streams RENAME COLUMN instances_made_primary_count TO instances_committed_count;
		`)
		if err != nil {
			return err
		}

		// StreamInstance
		_, err = db.Exec(`
			ALTER TABLE stream_instances DROP made_final_on;
			ALTER TABLE stream_instances RENAME COLUMN made_primary_on TO committed_on;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

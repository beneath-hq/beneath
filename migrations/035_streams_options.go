package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD use_log boolean NOT NULL DEFAULT true,
			ADD use_index boolean NOT NULL DEFAULT true,
			ADD use_warehouse boolean NOT NULL DEFAULT true;

			ALTER TABLE streams
			ADD log_retention_seconds integer NOT NULL DEFAULT 0,
			ADD index_retention_seconds integer NOT NULL DEFAULT 0,
			ADD warehouse_retention_seconds integer NOT NULL DEFAULT 0;

			UPDATE streams
			SET log_retention_seconds = retention_seconds,
					index_retention_seconds = retention_seconds,
					warehouse_retention_seconds = retention_seconds;

			ALTER TABLE streams
			DROP retention_seconds;

			ALTER TABLE streams
			RENAME COLUMN enable_manual_writes TO allow_manual_writes;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams RENAME COLUMN allow_manual_writes TO enable_manual_writes;

			ALTER TABLE streams
			ADD retention_seconds integer NOT NULL DEFAULT 0;

			ALTER TABLE streams
			DROP log_retention_seconds,
			DROP index_retention_seconds,
			DROP warehouse_retention_seconds;
			
			ALTER TABLE streams
			DROP use_log,
			DROP use_index,
			DROP use_warehouse;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

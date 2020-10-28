package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream.RetentionSeconds
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD retention_seconds integer NOT NULL DEFAULT 0;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Stream.RetentionSeconds
		_, err = db.Exec(`
			ALTER TABLE streams DROP retention_seconds;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

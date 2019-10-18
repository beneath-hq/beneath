package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream.InstancesCreatedCount, Stream.InstancesCommittedCount
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD instances_created_count integer NOT NULL DEFAULT 0,
			ADD instances_committed_count integer NOT NULL DEFAULT 0;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Stream.InstancesCreatedCount, Stream.InstancesCommittedCount
		_, err = db.Exec(`
			ALTER TABLE streams DROP instances_created_count;
			ALTER TABLE streams DROP instances_committed_count;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

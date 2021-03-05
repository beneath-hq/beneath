package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD next_instance_version bigint NOT NULL DEFAULT 0;
		`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
			UPDATE streams s
			SET    next_instance_version = sub.max_version + 1
			FROM (
				SELECT si.stream_id, max(si.version) as max_version
				FROM stream_instances si
				GROUP BY si.stream_id
			) AS sub
			WHERE sub.stream_id = s.stream_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams DROP next_instance_version;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

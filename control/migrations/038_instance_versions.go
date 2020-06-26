package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE stream_instances
			ADD version bigint NOT NULL default 0;

			CREATE UNIQUE INDEX ON public.stream_instances USING btree (stream_id, version);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE stream_instances
			DROP version;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

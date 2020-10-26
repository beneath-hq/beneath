package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD canonical_indexes JSON NOT NULL default '[]',
			DROP bigquery_schema;

			UPDATE streams s
			SET canonical_indexes = cast(replace(concat('[{"fields":', si.fields, ',"key":true}]'), ' ', '') as json)
			FROM stream_indexes si
			WHERE si.stream_id = s.stream_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			DROP canonical_indexes,
			ADD bigquery_schema JSON;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

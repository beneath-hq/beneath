package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// StreamIndex
		_, err = db.Exec(`
			CREATE TABLE stream_indexes(
				stream_index_id uuid DEFAULT uuid_generate_v4(),
				stream_id uuid NOT NULL,
				fields jsonb NOT NULL,
				"primary" boolean NOT NULL,
				normalize boolean NOT NULL,
				PRIMARY KEY (stream_index_id),
				FOREIGN KEY (stream_id) REFERENCES streams (stream_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		// (stream_id) index
		_, err = db.Exec(`
			CREATE INDEX ON public.stream_indexes (stream_id);
		`)
		if err != nil {
			return err
		}

		// migrate data
		_, err = db.Exec(`
			insert into stream_indexes(stream_id, fields, "primary", normalize)
			select s.stream_id, s.key_fields, true, true
			from streams s;
		`)
		if err != nil {
			return err
		}

		// drop key_fields
		_, err = db.Exec(`
			ALTER TABLE streams DROP key_fields;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// StreamIndex
		err = db.Model(&entity.StreamIndex{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

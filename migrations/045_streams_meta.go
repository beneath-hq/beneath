package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD meta boolean not null default false;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams
			DROP meta;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

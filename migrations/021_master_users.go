package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// User.Master
		_, err = db.Exec(`
			ALTER TABLE users
			ADD master boolean NOT NULL DEFAULT false;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// User.Master
		_, err = db.Exec(`
			ALTER TABLE users 
			DROP master;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

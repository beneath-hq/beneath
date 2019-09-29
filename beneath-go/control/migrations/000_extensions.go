package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Extensions
		_, err = db.Exec(`
			CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Extensions
		_, err = db.Exec(`
			DROP EXTENSION "uuid-ossp";
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

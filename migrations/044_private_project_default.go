package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Project.Public
		_, err = db.Exec(`
			ALTER TABLE projects 
			ALTER COLUMN public SET DEFAULT FALSE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Project.Public
		_, err = db.Exec(`
			ALTER TABLE projects 
			ALTER COLUMN public SET DEFAULT TRUE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

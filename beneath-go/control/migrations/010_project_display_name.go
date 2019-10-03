package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Project.DisplayName NOT NULL
		_, err = db.Exec(`
			ALTER TABLE projects ALTER COLUMN display_name DROP NOT NULL;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Project.DisplayName NOT NULL
		_, err = db.Exec(`
			ALTER TABLE projects ALTER COLUMN display_name SET NOT NULL;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

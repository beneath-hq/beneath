package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE projects
			ADD explore_rank integer;
			CREATE INDEX ON projects (explore_rank) WHERE explore_rank IS NOT NULL;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE projects
			DROP explore_rank;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

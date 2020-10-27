package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Service.CreatedOn and Service.UpdatedOn
		_, err = db.Exec(`
			ALTER TABLE services
			ADD created_on timestamp with time zone NOT NULL DEFAULT now(),
    	ADD updated_on timestamp with time zone NOT NULL DEFAULT now();
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Service.CreatedOn and Service.UpdatedOn
		_, err = db.Exec(`
			ALTER TABLE services DROP created_on;
			ALTER TABLE services DROP updated_on;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

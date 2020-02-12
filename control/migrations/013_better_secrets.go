package migrations

import (
	"github.com/beneath-core/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Secret
		_, err = db.Exec(`
			DROP TABLE secrets;
		`)
		if err != nil {
			return err
		}

		// ServiceSecret
		err = db.Model(&entity.ServiceSecret{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// UserSecret
		err = db.Model(&entity.UserSecret{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// not quite reversible due to dropped table

		err = db.Model(&entity.UserSecret{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		err = db.Model(&entity.ServiceSecret{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		return nil
	})
}

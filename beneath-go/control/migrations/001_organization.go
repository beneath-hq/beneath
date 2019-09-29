package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Organization
		err = db.Model(&entity.Organization{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Organization
		err = db.Model(&entity.Organization{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

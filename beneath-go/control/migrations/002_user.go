package migrations

import (
	"github.com/beneath-core/beneath-go/control/model"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// User
		err = db.Model(&model.User{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// UserToProject
		err = db.Model(&model.UserToProject{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// User
		err = db.Model(&model.User{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// UserToProject
		err = db.Model(&model.UserToProject{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

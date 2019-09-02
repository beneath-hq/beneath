package migrations

import (
	"github.com/beneath-core/beneath-go/control/model"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Secret
		err = db.Model(&model.Secret{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Secret
		err = db.Model(&model.Secret{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

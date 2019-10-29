package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Note: doesn't exist anymore
		// Secret
		// err = db.Model(&entity.Secret{}).CreateTable(defaultCreateOptions)
		// if err != nil {
		// 	return err
		// }

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Note: doesn't exist anymore
		// Secret
		// err = db.Model(&entity.Secret{}).DropTable(defaultDropOptions)
		// if err != nil {
		// 	return err
		// }

		// Done
		return nil
	})
}

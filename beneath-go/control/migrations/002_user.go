package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// User
		err = db.Model(&entity.User{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// User email index
		_, err = db.Exec(`
			ALTER TABLE users DROP CONSTRAINT users_email_key;
			CREATE UNIQUE INDEX users_email_key ON public.users USING btree ((lower(email)));
		`)
		if err != nil {
			return err
		}

		// ProjectToUser
		err = db.Model(&entity.ProjectToUser{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// User
		err = db.Model(&entity.User{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// ProjectToUser
		err = db.Model(&entity.ProjectToUser{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

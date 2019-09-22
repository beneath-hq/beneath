package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Service
		err = db.Model(&entity.Service{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// PermissionsServicesStreams
		err = db.Model(&entity.PermissionsServicesStreams{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Service
		err = db.Model(&entity.Service{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// PermissionsServicesStreams
		err = db.Model(&entity.PermissionsServicesStreams{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

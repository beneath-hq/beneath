package migrations

import (
	"github.com/beneath-core/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Stream
		_, err = db.Exec(`
			CREATE TABLE services (
				service_id uuid DEFAULT uuid_generate_v4(),
				name text NOT NULL,
				kind text NOT NULL,
				organization_id uuid NOT NULL,
				read_quota bigint,
				write_quota bigint,
				PRIMARY KEY (service_id),
				FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE
			);
		`)
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

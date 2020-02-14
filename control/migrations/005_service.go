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
				name text,
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
		_, err = db.Exec(`
			CREATE TABLE permissions_services_streams (
				service_id uuid,
				stream_id uuid,
				read boolean NOT NULL,
				write boolean NOT NULL,
				PRIMARY KEY (service_id, stream_id),
				FOREIGN KEY (service_id) REFERENCES services (service_id) ON DELETE CASCADE,
				FOREIGN KEY (stream_id) REFERENCES streams (stream_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// PermissionsServicesStreams
		err = db.Model(&entity.PermissionsServicesStreams{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Service
		err = db.Model(&entity.Service{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

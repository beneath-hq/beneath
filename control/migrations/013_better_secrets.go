package migrations

import (
	"gitlab.com/beneath-hq/beneath/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Secret
		_, err = db.Exec(`
			DROP TABLE IF EXISTS secrets;
		`)
		if err != nil {
			return err
		}

		// ServiceSecret
		_, err = db.Exec(`
			CREATE TABLE service_secrets (
				service_secret_id uuid DEFAULT uuid_generate_v4(),
				prefix text NOT NULL,
				hashed_token bytea NOT NULL UNIQUE,
				description text,
				created_on timestamptz NOT NULL DEFAULT now(),
				updated_on timestamptz NOT NULL DEFAULT now(),
				service_id uuid,
				PRIMARY KEY (service_secret_id),
				FOREIGN KEY (service_id) REFERENCES services (service_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		// UserSecret
		_, err = db.Exec(`
			CREATE TABLE user_secrets (
				user_secret_id uuid DEFAULT uuid_generate_v4(),
				prefix text NOT NULL,
				hashed_token bytea NOT NULL UNIQUE,
				description text,
				created_on timestamptz NOT NULL DEFAULT now(),
				updated_on timestamptz NOT NULL DEFAULT now(),
				user_id uuid NOT NULL,
				read_only boolean NOT NULL,
				public_only boolean NOT NULL,
				PRIMARY KEY (user_secret_id),
				FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
			);
		`)
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

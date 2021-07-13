package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			CREATE TABLE auth_tickets (
				auth_ticket_id uuid DEFAULT uuid_generate_v4(),
				requester_name text NOT NULL,
				approver_user_id uuid,
				created_on timestamptz NOT NULL DEFAULT now(),
				updated_on timestamptz NOT NULL DEFAULT now(),
				PRIMARY KEY (auth_ticket_id),
				FOREIGN KEY (approver_user_id) REFERENCES users(user_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			DROP TABLE IF EXISTS auth_tickets;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

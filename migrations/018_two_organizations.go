package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// User.BillingOrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN organization_id TO billing_organization_id;
		`)
		if err != nil {
			return err
		}

		// User.PersonalOrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			ADD personal_organization_id uuid REFERENCES organizations (organization_id) ON DELETE restrict;
		`)
		if err != nil {
			return err
		}

		// (billing_organization_id) index
		_, err = db.Exec(`
			CREATE INDEX users_billing_organization_id_key ON public.users USING btree (billing_organization_id);
		`)
		if err != nil {
			return err
		}

		// (personal_organization_id) index
		_, err = db.Exec(`
			CREATE INDEX users_personal_organization_id_key ON public.users USING btree (personal_organization_id);
		`)
		if err != nil {
			return err
		}

		// migrate data
		_, err = db.Exec(`
			insert into users(personal_organization_id)
			select u.billing_organization_id
			from users u;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// (billing_organization_id) index
		_, err = db.Exec(`
			DROP INDEX users_billing_organization_id_key;
		`)
		if err != nil {
			return err
		}

		// User.PersonalOrganizationID
		_, err = db.Exec(`
			ALTER TABLE users DROP personal_organization_id;
		`)
		if err != nil {
			return err
		}

		// User.BillingOrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN billing_organization_id TO organization_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

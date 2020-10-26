package migrations

import (
	"gitlab.com/beneath-hq/beneath/control/entity"

	"github.com/go-pg/migrations/v7"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// OrganizationInvite
		_, err = db.Exec(`
			CREATE TABLE organization_invites (
				organization_invite_id uuid DEFAULT uuid_generate_v4(),
				organization_id uuid NOT NULL,
				user_id uuid NOT NULL,
				created_on timestamptz NOT NULL DEFAULT now(),
				updated_on timestamptz NOT NULL DEFAULT now(),
				view boolean NOT NULL,
				"create" boolean NOT NULL,
				admin boolean NOT NULL,
				PRIMARY KEY (organization_invite_id),
				FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
			);
		`)
		if err != nil {
			return err
		}

		// Only one invitation per (organization_id, user_id)
		_, err = db.Exec(`
			CREATE UNIQUE INDEX ON public.organization_invites USING btree (organization_id, user_id);
		`)
		if err != nil {
			return err
		}

		// Index on user_id
		_, err = db.Exec(`
			CREATE INDEX ON public.organization_invites (user_id);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// OrganizationInvite
		err = db.Model(&entity.OrganizationInvite{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

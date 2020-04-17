package migrations

import (
	"github.com/go-pg/migrations/v7"

	"gitlab.com/beneath-hq/beneath/control/entity"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan
		err = db.Model(&entity.BillingPlan{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// BilledResource
		err = db.Model(&entity.BilledResource{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// BillingInfo
		_, err = db.Exec(`
			CREATE TABLE billing_infos(
				billing_info_id       UUID DEFAULT uuid_generate_v4(),
				organization_id				UUID NOT NULL,
				billing_plan_id 			UUID NOT NULL,
				payments_driver 			TEXT NOT NULL,
				driver_payload 				JSONB,
				created_on 						timestamp with time zone DEFAULT now(),
				updated_on 						timestamp with time zone DEFAULT now(),
				PRIMARY KEY (billing_info_id),
				FOREIGN KEY (organization_id) REFERENCES organizations (organization_id) ON DELETE CASCADE,
				FOREIGN KEY (billing_plan_id) REFERENCES billing_plans (billing_plan_id) ON DELETE RESTRICT
			);
		`)
		if err != nil {
			return err
		}

		// only one default BillingPlan
		_, err = db.Exec(`
			CREATE UNIQUE INDEX ON billing_plans ("default") WHERE "default" = true;
		`)
		if err != nil {
			return err
		}

		// (billing_time, org_id, entity_id, product) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX billed_resources_billing_time_organization_id_entity_id_product_key
				ON public.billed_resources USING btree (billing_time, organization_id, entity_id, product);
		`)
		if err != nil {
			return err
		}

		// Organization.Personal
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD personal bool NOT NULL default FALSE;
		`)
		if err != nil {
			return err
		}

		// User.OrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN main_organization_id TO organization_id;
		`)
		if err != nil {
			return err
		}

		// Project.Locked
		_, err = db.Exec(`
			ALTER TABLE projects ADD locked bool NOT NULL default FALSE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {

		// Project.Locked
		_, err = db.Exec(`
			ALTER TABLE projects DROP locked;
		`)
		if err != nil {
			return err
		}

		// User.OrganizationID
		_, err = db.Exec(`
			ALTER TABLE users
			RENAME COLUMN organization_id TO main_organization_id;
		`)
		if err != nil {
			return err
		}

		// Organization.Personal
		_, err = db.Exec(`
			ALTER TABLE organizations DROP personal;
		`)
		if err != nil {
			return err
		}

		// BillingInfo
		err = db.Model(&entity.BillingInfo{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// BilledResource
		err = db.Model(&entity.BilledResource{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// BillingPlan
		err = db.Model(&entity.BillingPlan{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

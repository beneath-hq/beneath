package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations/v7"
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
		err = db.Model(&entity.BillingInfo{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// (billing_time, org_id, entity_id, product) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX billed_resources_billing_time_organization_id_entity_id_product_key ON public.billed_resources USING btree (billing_time, organization_id, entity_id, product);
		`)
		if err != nil {
			return err
		}

		// Organization.Personal, Organization.BillingPlanID, Organization.StripeCustomerID, Organization.PaymentMethod
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

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan
		err = db.Model(&entity.BillingPlan{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// BilledResource
		err = db.Model(&entity.BilledResource{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// BillingInfo
		err = db.Model(&entity.BillingInfo{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Organization.Personal, Organization.BillingPlanID, Organization.StripeCustomerID, Organization.PaymentMethod
		_, err = db.Exec(`
			ALTER TABLE organizations DROP personal;
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

		// Done
		return nil
	})
}

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
			ADD personal bool NOT NULL default FALSE,
			ADD billing_plan_id uuid,
			ADD stripe_customer_id varchar,
			ADD payment_method varchar;
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

		// Organization.Personal, Organization.BillingPlanID, Organization.StripeCustomerID, Organization.PaymentMethod
		_, err = db.Exec(`
			ALTER TABLE organizations DROP personal;
			ALTER TABLE organizations DROP billing_plan_id;
			ALTER TABLE organizations DROP stripe_customer_id;
			ALTER TABLE organizations DROP payment_method;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

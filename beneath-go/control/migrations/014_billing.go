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

		// Organization.Personal, Organization.BillingPlanID, Organization.StripeCustomerID
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD personal bool NOT NULL default FALSE,
			ADD billing_plan_id uuid,
			ADD stripe_customer_id varchar;
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

		// Organization.Personal, Organization.BillingPlanID, Organization.StripeCustomerID
		_, err = db.Exec(`
			ALTER TABLE organizations DROP personal;
			ALTER TABLE organizations DROP billing_plan_id;
			ALTER TABLE organizations DROP stripe_customer_id;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

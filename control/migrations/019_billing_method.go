package migrations

import (
	"github.com/go-pg/migrations/v7"
	"gitlab.com/beneath-hq/beneath/control/entity"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingMethod
		err = db.Model(&entity.BillingMethod{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// BillingInfo.BillingMethodID
		_, err = db.Exec(`
			ALTER TABLE billing_infos
			ADD billing_method_id UUID NOT NULL,
			ADD FOREIGN KEY (billing_method_id) REFERENCES billing_methods (billing_method_id) ON DELETE RESTRICT;
		`)
		if err != nil {
			return err
		}

		// BillingInfo.PaymentsDriver and BillingInfo.DriverPayload
		_, err = db.Exec(`
			ALTER TABLE billing_infos DROP payments_driver;
			ALTER TABLE billing_infos DROP driver_payload;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingInfo.PaymentsDriver and BillingInfo.DriverPayload
		_, err = db.Exec(`
			ALTER TABLE billing_infos ADD payments_driver text;
			ALTER TABLE billing_infos ADD driver_payload jsonb;
		`)
		if err != nil {
			return err
		}

		// BillingInfo.BillingMethodID
		_, err = db.Exec(`
			ALTER TABLE billing_infos DROP billing_method_id;
		`)
		if err != nil {
			return err
		}

		// BillingMethod
		err = db.Model(&entity.BillingMethod{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

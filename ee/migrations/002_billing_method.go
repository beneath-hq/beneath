package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingMethod
		_, err = db.Exec(`
			CREATE TABLE billing_methods (
				billing_method_id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
				organization_id uuid NOT NULL REFERENCES organizations(organization_id) ON DELETE CASCADE,
				payments_driver text NOT NULL,
				driver_payload jsonb,
				created_on timestamp with time zone DEFAULT now(),
				updated_on timestamp with time zone DEFAULT now()
			);
		`)
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

		_, err = db.Exec(`
			DROP TABLE IF EXISTS billing_methods CASCADE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

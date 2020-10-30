package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan.BaseScanQuota, BillingPlan.SeatScanQuota, BillingPlan.ScanQuota, BillingPlan.ScanOveragePriceCents
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			ADD base_scan_quota bigint NOT NULL default 0,
			ADD seat_scan_quota bigint NOT NULL default 0,
			ADD scan_quota bigint NOT NULL default 0,
			ADD scan_overage_price_cents integer NOT NULL default 0;
		`)
		if err != nil {
			return err
		}

		// set values for Free plan
		_, err = db.Exec(`
			update billing_plans
			set base_scan_quota = 100000000000,
					scan_quota = 100000000000
			where description = 'Free';
		`)
		if err != nil {
			return err
		}

		// set values for Professional plan
		_, err = db.Exec(`
			update billing_plans
			set base_scan_quota = 1000000000000,
					scan_quota = 1000000000000,
					scan_overage_price_cents = 4
			where description = 'Professional';
		`)
		if err != nil {
			return err
		}

		// Organizations.ScanQuota, Organizations.PrepaidScanQuota
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD scan_quota bigint,
			ADD prepaid_scan_quota bigint;
		`)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
			update organizations
			set scan_quota = 100000000000,
					prepaid_scan_quota = 100000000000;
		`)
		if err != nil {
			return err
		}

		// Users.ScanQuota
		_, err = db.Exec(`
			ALTER TABLE users
			ADD scan_quota bigint;
		`)
		if err != nil {
			return err
		}

		// Services.ScanQuota
		_, err = db.Exec(`
			ALTER TABLE services
			ADD scan_quota bigint;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Services.ScanQuota
		_, err = db.Exec(`
			ALTER TABLE services
			DROP scan_quota;
		`)
		if err != nil {
			return err
		}

		// Users.ScanQuota
		_, err = db.Exec(`
			ALTER TABLE users
			DROP scan_quota;
		`)
		if err != nil {
			return err
		}

		// Organizations.ScanQuota, Organizations.PrepaidScanQuota
		_, err = db.Exec(`
			ALTER TABLE organizations
			DROP scan_quota,
			DROP prepaid_scan_quota;
		`)
		if err != nil {
			return err
		}

		// BillingPlan.BaseScanQuota, BillingPlan.SeatScanQuota, BillingPlan.ScanQuota, BillingPlan.ScanOveragePriceCents
		_, err = db.Exec(`
			ALTER TABLE billing_plans
			DROP base_scan_quota,
			DROP seat_scan_quota,
			DROP scan_quota,
			DROP scan_overage_price_cents;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

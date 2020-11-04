package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
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

		return nil
	})
}

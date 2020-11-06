package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE organizations ADD quota_epoch timestamptz;
			ALTER TABLE services ADD quota_epoch timestamptz;
			ALTER TABLE users ADD quota_epoch timestamptz;
			UPDATE organizations o SET quota_epoch = date_trunc('month', o.created_on);
			UPDATE services s SET quota_epoch = date_trunc('month', s.created_on);
			UPDATE users u SET quota_epoch = date_trunc('month', u.created_on);
		`)
		if err != nil {
			return err
		}

		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE organizations DROP quota_epoch;
			ALTER TABLE services DROP quota_epoch;
			ALTER TABLE users DROP quota_epoch;
		`)
		if err != nil {
			return err
		}

		return nil
	})
}

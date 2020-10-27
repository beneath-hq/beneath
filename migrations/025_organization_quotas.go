package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Organizations.ReadQuota, Organizations.WriteQuota
		_, err = db.Exec(`
			ALTER TABLE organizations
			ADD read_quota bigint,
			ADD write_quota bigint;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Organizations.ReadQuota, Organizations.WriteQuota
		_, err = db.Exec(`
			ALTER TABLE organizations 
			DROP read_quota,
			DROP write_quota;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

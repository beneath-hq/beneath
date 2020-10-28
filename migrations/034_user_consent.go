package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// User.ConsentTerms, User.ConsentNewsletter
		_, err = db.Exec(`
			ALTER TABLE users
			ADD consent_terms boolean NOT NULL DEFAULT false,
			ADD consent_newsletter boolean NOT NULL DEFAULT false;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// BillingPlan.MultipleUsers
		_, err = db.Exec(`
			ALTER TABLE users
			DROP consent_terms,
			DROP consent_newsletter;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}

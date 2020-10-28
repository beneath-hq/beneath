package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// Removed
		return nil
	}, func(db migrations.DB) (err error) {
		// Removed
		return nil
	})
}

package migrations

import (
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/go-pg/migrations/v7"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

var (
	defaultCreateOptions *orm.CreateTableOptions
	defaultDropOptions   *orm.DropTableOptions
)

func init() {
	defaultCreateOptions = &orm.CreateTableOptions{
		FKConstraints: true,
	}

	defaultDropOptions = &orm.DropTableOptions{
		IfExists: false,
		Cascade:  true,
	}
}

// Run forwards args to https://godoc.org/github.com/go-pg/migrations/v7#Run
func Run(db *pg.DB, a ...string) (oldVersion, newVersion int64, err error) {
	return migrations.Run(db, a...)
}

// MustRunUp initializes migrations (if necessary) then applies all new migrations; it panics on error
func MustRunUp(db *pg.DB) {
	// disable searching for sql files (after go build, migrations folder no longer exists in Docker image)
	migrations.DefaultCollection.DisableSQLAutodiscover(true)

	// init migrations if not already initialized
	_, _, err := migrations.Run(db, "init")
	if err != nil {
		// ignore error if migration table already exists
		if err.Error() != "ERROR #42P07 relation \"gopg_migrations\" already exists" {
			log.S.Fatalf("migrations: %s", err.Error())
		}
	}

	// run migrations
	oldVersion, newVersion, err := migrations.Run(db, "up")
	if err != nil {
		log.S.Fatalf("migrations: %s", err.Error())
	}

	// log version status
	if newVersion != oldVersion {
		log.S.Infof("migrated from version %d to %d", oldVersion, newVersion)
	} else {
		log.S.Infof("running at migration version %d", oldVersion)
	}
}

package migrations

import (
	"gitlab.com/beneath-hq/beneath/pkg/migrationsutil"
)

// Migrator registers and runs migrations
var Migrator = migrationsutil.New("gopg_migrations_ee")

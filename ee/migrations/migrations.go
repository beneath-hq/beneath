package migrations

import (
	"github.com/beneath-hq/beneath/pkg/migrationsutil"
)

// Migrator registers and runs migrations.
// Tracks migrations in a different table than non-enterprise migrations.
var Migrator = migrationsutil.New("gopg_migrations_ee")

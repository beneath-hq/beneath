package migrationsutil

import (
	"fmt"

	"github.com/go-pg/migrations/v7"
	"github.com/go-pg/pg/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/pkg/envutil"
)

// Migrator wraps go-pg/migrations with useful extra functionality
type Migrator struct {
	*migrations.Collection
	TableName string
}

// New creates a new migrator
func New(tableName string) *Migrator {
	collection := migrations.NewCollection()
	if tableName != "" {
		collection.SetTableName(tableName)
	}
	collection.DisableSQLAutodiscover(true)
	return &Migrator{
		Collection: collection,
		TableName:  tableName,
	}
}

// AddCmd registers a migration CLI with Cobra.
// The callback carries args that can be supplied outright to RunWithArgs.
func (m *Migrator) AddCmd(root *cobra.Command, name string, fn func(args []string)) {
	runWithoutName := func(cmd *cobra.Command, args []string) { fn(args) }
	runWithName := func(cmd *cobra.Command, args []string) { fn(append([]string{cmd.Name()}, args...)) }

	cmd := &cobra.Command{
		Use:   name,
		Short: "Manages migrations on the db (runs \"up\" if no subcommand provided)",
		Args:  cobra.NoArgs,
		Run:   runWithoutName,
	}
	root.AddCommand(cmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Creates version info table in the database",
		Args:  cobra.NoArgs,
		Run:   runWithName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "up [target]",
		Short: "Runs available migrations up to the target (or all if target not set)",
		Args:  cobra.MaximumNArgs(1),
		Run:   runWithName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Reverts last migration",
		Args:  cobra.NoArgs,
		Run:   runWithName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reverts all migrations",
		Args:  cobra.NoArgs,
		Run:   runWithName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Prints latest migration version",
		Args:  cobra.NoArgs,
		Run:   runWithName,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set_version [version]",
		Short: "",
		Args:  cobra.ExactArgs(1),
		Run:   runWithName,
	})
}

// RunWithArgs runs migrations with command-line args (as registered with RegisterCmd)
func (m *Migrator) RunWithArgs(db *pg.DB, logger *zap.Logger, args ...string) {
	oldVersion, newVersion, err := m.Collection.Run(db, args...)
	m.log(logger, oldVersion, newVersion, err)
}

// Automigrate automatically applies every new migration. If reset is true, it resets migrations
// before automigrating.
func (m *Migrator) Automigrate(db *pg.DB, reset bool) (oldVersion, newVersion int64, err error) {
	// safety check on reset in production
	if reset && envutil.GetEnv() == envutil.Production {
		return 0, 0, fmt.Errorf("Cannot automigrate with reset=true in production")
	}

	// run init only if necessary
	_, err = m.Version(db)
	if err != nil {
		if err.Error() != fmt.Sprintf(`ERROR #42P01 relation "%s" does not exist`, m.TableName) {
			return 0, 0, err
		}

		_, _, err := m.Collection.Run(db, "init")
		if err != nil {
			return 0, 0, err
		}
	}

	// run reset if requested
	if reset {
		_, _, err := m.Collection.Run(db, "reset")
		if err != nil {
			return 0, 0, err
		}
	}

	// run up migrations
	oldVersion, newVersion, err = m.Collection.Run(db, "up")
	if err != nil {
		return oldVersion, newVersion, err
	}

	return oldVersion, newVersion, err
}

// AutomigrateAndLog runs m.Automigrate and logs the result, returning only an error if applicable
func (m *Migrator) AutomigrateAndLog(db *pg.DB, logger *zap.Logger, reset bool) error {
	oldVersion, newVersion, err := m.Automigrate(db, reset)
	m.log(logger, oldVersion, newVersion, err)
	return err
}

func (m *Migrator) log(logger *zap.Logger, oldVersion, newVersion int64, err error) {
	l := logger.Named("migrator")
	if err != nil {
		l.Sugar().Errorf("failed running migrations on '%s': %s", m.TableName, err.Error())
	} else if newVersion != oldVersion {
		l.Sugar().Infof("migrated '%s' from version %d to %d", m.TableName, oldVersion, newVersion)
	} else {
		l.Sugar().Infof("migration version is %d on '%s'", newVersion, m.TableName)
	}
}

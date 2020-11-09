package main

import (
	"context"

	"go.uber.org/zap"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/migrations"
	"gitlab.com/beneath-hq/beneath/server/control"
	"gitlab.com/beneath-hq/beneath/server/data"
	"gitlab.com/beneath-hq/beneath/services/usage"

	// registers all dependencies with the CLI
	_ "gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
)

func main() {
	cli := cli.NewCLI()
	addMigrateCmd(cli)
	cli.Run()
}

// registers migrate command
func addMigrateCmd(c *cli.CLI) {
	migrations.Migrator.AddCmd(c.Root, "migrate", func(args []string) {
		cli.Dig.Invoke(func(db db.DB, logger *zap.Logger) {
			migrations.Migrator.RunWithArgs(db, logger, args...)
		})
	})
}

func init() {
	cli.AddStartable(&cli.Startable{
		Name: "control-server",
		Register: func(lc *cli.Lifecycle, server *control.Server, db db.DB, logger *zap.Logger) {
			// running the control server also runs automigrate
			lc.AddFunc("automigrate", func(ctx context.Context) error {
				return migrations.Migrator.AutomigrateAndLog(db, logger, false)
			})

			// add control server to lifecycle
			lc.Add("control-server", server)
		},
	})

	cli.AddStartable(&cli.Startable{
		Name: "data-server",
		Register: func(lc *cli.Lifecycle, server *data.Server, usage *usage.Service) {
			lc.Add("data-server", server)
			lc.Add("usage-writer", usage)
		},
	})

	cli.AddStartable(&cli.Startable{
		Name: "control-worker",
		Register: func(lc *cli.Lifecycle, worker *dependencies.ControlWorker) {
			lc.Add("control-worker", worker)
		},
	})

	cli.AddStartable(&cli.Startable{
		Name: "data-worker",
		Register: func(lc *cli.Lifecycle, worker *dependencies.DataWorker) {
			lc.Add("data-worker", worker)
		},
	})
}

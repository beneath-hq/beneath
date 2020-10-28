package main

import (
	"context"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	eedependencies "gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	eemigrations "gitlab.com/beneath-hq/beneath/ee/migrations"
	eecontrol "gitlab.com/beneath-hq/beneath/ee/server/control"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/migrations"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/server/control"
	"gitlab.com/beneath-hq/beneath/server/data"
	"gitlab.com/beneath-hq/beneath/services/metrics"

	// registers all dependencies with the CLI
	_ "gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	_ "gitlab.com/beneath-hq/beneath/ee/cmd/beneath/dependencies"
)

func main() {
	cli := cli.NewCLI()
	addMigrateCmd(cli)
	cli.Run()
}

// registers migrate command
func addMigrateCmd(c *cli.CLI) {
	log.InitLogger()
	migrations.Migrator.AddCmd(c.Root, "migrate", func(args []string) {
		cli.Dig.Invoke(func(db db.DB) {
			migrations.Migrator.RunWithArgs(db, args...)
		})
	})
	migrations.Migrator.AddCmd(c.Root, "migrate-ee", func(args []string) {
		cli.Dig.Invoke(func(db db.DB) {
			eemigrations.Migrator.RunWithArgs(db, args...)
		})
	})
}

func init() {
	cli.AddStartable(&cli.Startable{
		Name: "control-server",
		Register: func(lc *cli.Lifecycle, server *control.Server, eeServer *eecontrol.Server, db db.DB) {
			// running the control server also runs automigrate
			lc.AddFunc("automigrate", func(ctx context.Context) error {
				return migrations.Migrator.AutomigrateAndLog(db, false)
			})
			lc.AddFunc("automigrate-ee", func(ctx context.Context) error {
				return eemigrations.Migrator.AutomigrateAndLog(db, false)
			})

			// mounts the enterprise control server on the "/ee" route of the normal control server (parasite!)
			server.Router.Mount("/ee", eeServer.Router)

			// add control server to lifecycle
			lc.Add("control-server", server)
		},
	})

	cli.AddStartable(&cli.Startable{
		Name: "data-server",
		Register: func(lc *cli.Lifecycle, server *data.Server, metrics *metrics.Broker) {
			lc.Add("data-server", server)
			lc.Add("metrics-broker", metrics)
		},
	})

	cli.AddStartable(&cli.Startable{
		Name: "control-worker",
		Register: func(lc *cli.Lifecycle, worker *dependencies.ControlWorker, eeServices *eedependencies.AllServices) {
			// eeServices is a dependency to ensure that all event listeners get registered
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

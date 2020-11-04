package main

import (
	"context"

	"github.com/spf13/cobra"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	eedependencies "gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
	eemigrations "gitlab.com/beneath-hq/beneath/ee/migrations"
	eecontrol "gitlab.com/beneath-hq/beneath/ee/server/control"
	"gitlab.com/beneath-hq/beneath/ee/services/billing"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/migrations"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/server/control"
	"gitlab.com/beneath-hq/beneath/server/data"
	"gitlab.com/beneath-hq/beneath/services/usage"

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

func addBillingCmd(c *cli.CLI) {
	log.InitLogger()
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing-related functionality",
	}
	billingCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Runs billing for all customers whose plan is set to renew since the last invocation",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cli.Dig.Invoke(func(billing *billing.Service) {
				err := billing.RunBilling(context.Background())
				if err != nil {
					log.S.Errorf("Billing failed with error: %s", err.Error())
				}
			})
		},
	})
	c.Root.AddCommand(billingCmd)
}

func init() {
	cli.AddStartable(&cli.Startable{
		Name: "control-server",
		Register: func(lc *cli.Lifecycle, server *control.Server, eeServer *eecontrol.Server, db db.DB) {
			// running the control server also runs automigrate (ce migrations first, then ee migrations)
			lc.AddFunc("automigrate", func(ctx context.Context) error {
				err := migrations.Migrator.AutomigrateAndLog(db, false)
				if err != nil {
					return err
				}
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
		Register: func(lc *cli.Lifecycle, server *data.Server, usage *usage.Service) {
			lc.Add("data-server", server)
			lc.Add("usage-writer", usage)
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

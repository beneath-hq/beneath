package main

import (
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/migrations"
	"gitlab.com/beneath-hq/beneath/pkg/log"

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
	log.InitLogger()
	migrations.Migrator.AddCmd(c.Root, "migrate", func(args []string) {
		cli.Dig.Invoke(func(db db.DB) {
			migrations.Migrator.RunWithArgs(db, args...)
		})
	})
}

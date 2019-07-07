// Command line tool to run/rollback migrations
// Essentially stolen from https://github.com/go-pg/migrations/blob/master/example/main.go

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/control/migrations"
	"github.com/beneath-core/beneath-go/core"
)

const usageText = `This program runs command on the db. Supported commands are:
  - init - creates version info table in the database
  - up - runs all available migrations.
  - up [target] - runs available migrations up to the target one.
  - down - reverts last migration.
  - reset - reverts all migrations.
  - version - prints current db version.
  - set_version [version] - sets db version without running migrations.
Usage:
  go run *.go <command> [args]
`

type configSpecification struct {
	PostgresURL string `envconfig:"POSTGRES_URL" required:"true"`
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var config configSpecification
	core.LoadConfig("beneath", &config)
	db.InitPostgres(config.PostgresURL)

	oldVersion, newVersion, err := migrations.Run(db.DB, flag.Args()...)
	if err != nil {
		exitf(err.Error())
	}

	if newVersion != oldVersion {
		fmt.Printf("migrated from version %d to %d\n", oldVersion, newVersion)
	} else {
		fmt.Printf("version is %d\n", oldVersion)
	}
}

func usage() {
	fmt.Print(usageText)
	flag.PrintDefaults()
	os.Exit(2)
}

func errorf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

func exitf(s string, args ...interface{}) {
	errorf(s, args...)
	os.Exit(1)
}

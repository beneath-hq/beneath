package main

import (
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"

	// registers everything all dependencies with the CLI
	_ "gitlab.com/beneath-hq/beneath/cmd/beneath/dependencies"
)

func main() {
	cli.NewCLI().Run()
}

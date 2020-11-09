package dependencies

import (
	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
)

func init() {
	cli.AddDependency(bus.NewBus)
}

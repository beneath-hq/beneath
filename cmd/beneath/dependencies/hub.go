package dependencies

import (
	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
)

func init() {
	cli.AddDependency(bus.NewBus)
}

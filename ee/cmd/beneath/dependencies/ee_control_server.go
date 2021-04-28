package dependencies

import (
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	eecontrol "github.com/beneath-hq/beneath/ee/server/control"
)

func init() {
	cli.AddDependency(eecontrol.NewServer)
}

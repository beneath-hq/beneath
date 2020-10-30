package dependencies

import (
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	eecontrol "gitlab.com/beneath-hq/beneath/ee/server/control"
)

func init() {
	cli.AddDependency(eecontrol.NewServer)
}

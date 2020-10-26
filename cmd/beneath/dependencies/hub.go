package dependencies

import (
	"context"

	"github.com/go-pg/pg/v9"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
)

func init() {
	cli.AddDependency(bus.NewBus)

	cli.AddDependency(func(db db.DB) *pg.DB {
		return db.GetDB(context.Background()).(*pg.DB)
	})
}

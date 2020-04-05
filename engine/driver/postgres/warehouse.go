package postgres

import (
	"context"

	"gitlab.com/beneath-org/beneath/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (b Postgres) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	panic("todo")
}

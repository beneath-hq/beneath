package postgres

import (
	"context"

	"gitlab.com/beneath-hq/beneath/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (b Postgres) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	panic("todo")
}

// ReadWarehouseCursor implements beneath.WarehouseService
func (b Postgres) ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

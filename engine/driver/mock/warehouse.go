package mock

import (
	"context"

	"gitlab.com/beneath-hq/beneath/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (m Mock) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	return nil
}

// ReadWarehouseCursor implements beneath.WarehouseService
func (m Mock) ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (driver.RecordsIterator, error) {
	return mockRecordsIterator{}, nil
}

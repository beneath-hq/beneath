package mock

import (
	"context"

	"gitlab.com/beneath-org/beneath/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (m Mock) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	return nil
}

package mock

import (
	"context"

	"github.com/beneath-core/engine/driver"
)

// WriteToWarehouse implements beneath.WarehouseService
func (m Mock) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	return nil
}

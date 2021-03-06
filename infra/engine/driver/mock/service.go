package mock

import (
	"context"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

// MaxKeySize implements beneath.Service
func (m Mock) MaxKeySize() int {
	return 1024 // 1 kb
}

// MaxRecordSize implements beneath.Service
func (m Mock) MaxRecordSize() int {
	return 1048576 // 1 mb
}

// MaxRecordsInBatch implements beneath.Service
func (m Mock) MaxRecordsInBatch() int {
	return 10000
}

// RegisterInstance implements beneath.Service
func (m Mock) RegisterInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	return nil
}

// RemoveInstance implements beneath.Service
func (m Mock) RemoveInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	return nil
}

// Reset implements beneath.Service
func (m Mock) Reset(ctx context.Context) error {
	return nil
}

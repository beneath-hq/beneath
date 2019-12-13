package engine

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// ReadRecords reads limit records starting at the given offset
func (e *Engine) ReadRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, offset int, limit int) (driver.RecordsReader, error) {
	return e.Log.ReadRecords(ctx, p, s, i, offset, limit)
}

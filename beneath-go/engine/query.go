package engine

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// Query returns cursors for paging through data under the given query
func (e *Engine) Query(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query, compact bool, partitions int) (replayCursors, changeCursors [][]byte, err error) {
	panic("not implemented")
	// return e.Log.ParseLogQuery(ctx, p, s, i, where)
}

// ReadCursor returns records starting at the cursor
func (e *Engine) ReadCursor(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("not implemented")
}

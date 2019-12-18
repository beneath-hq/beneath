package engine

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// QueryLog returns a cursor for paging through the log under the given query
func (e *Engine) QueryLog(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query) ([]byte, error) {
	return e.Log.ParseLogQuery(ctx, p, s, i, where)
}

// ReadLog reads rows from the log starting at cursor
func (e *Engine) ReadLog(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	return e.Log.ReadLog(ctx, p, s, i, cursor, limit)
}

package engine

import (
	"context"

	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
)

// QueryLookup returns a cursor for efficiently paging through records satisfying the where query
func (e *Engine) QueryLookup(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query) ([]byte, error) {
	panic("todo")
}

// ReadLookup returns ...
func (e *Engine) ReadLookup(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursor []byte, limit int) (driver.RecordsIterator, error) {
	panic("todo")
}

package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"gitlab.com/beneath-hq/beneath/infra/engine/driver"
)

const (
	maxReferencedStreams = 5
)

var (
	warehouseQueryStreamRegex = regexp.MustCompile("`(/?[_\\-a-zA-Z0-9]+/[_:\\-/a-zA-Z0-9]+)`")
)

// WarehouseQueryStreamResolver is a callback that should resolve a stream qualifier
type WarehouseQueryStreamResolver func(ctx context.Context, organization, project, stream string) (driver.Project, driver.Stream, driver.StreamInstance, error)

// ExpandWarehouseQuery replaces stream references of the form `org/proj/stream` in warehouse queries
// with the proper name of the underlying table in the warehouse
func (e *Engine) ExpandWarehouseQuery(ctx context.Context, query string, resolver WarehouseQueryStreamResolver) (string, error) {
	// find stream paths in query
	matches := warehouseQueryStreamRegex.FindAllString(query, -1)
	if len(matches) == 0 {
		return query, nil
	}

	// create mapping of streams to resolve
	mu := &sync.Mutex{}
	tableNames := make(map[streamQualifier]string)
	for _, path := range matches {
		qualifier, err := newStreamQualifier(path)
		if err != nil {
			return "", err
		}
		tableNames[qualifier] = ""
	}

	// check number of tables is within limit
	if len(tableNames) > maxReferencedStreams {
		return "", fmt.Errorf("a query cannot reference more than %d streams", maxReferencedStreams)
	}

	// use an errgroup to concurrently lookup all the streams referenced in the query
	group, gctx := errgroup.WithContext(ctx)
	for qualifier := range tableNames {
		qualifier := qualifier // bind locally
		group.Go(func() error {
			p, s, si, err := resolver(gctx, qualifier.Organization, qualifier.Project, qualifier.Stream)
			if err != nil {
				return err
			}

			tableName := e.Warehouse.GetWarehouseTableName(p, s, si)
			mu.Lock()
			tableNames[qualifier] = tableName
			mu.Unlock()

			return nil
		})
	}
	err := group.Wait()
	if err != nil {
		return "", err
	}

	// execute actual expand
	expandedQuery := warehouseQueryStreamRegex.ReplaceAllStringFunc(query, func(path string) string {
		qualifier, err := newStreamQualifier(path)
		if err != nil {
			panic(err) // shouldn't happen; would have been caught earlier
		}
		return tableNames[qualifier]
	})

	return expandedQuery, nil
}

type streamQualifier struct {
	Organization string
	Project      string
	Stream       string
}

func newStreamQualifier(path string) (streamQualifier, error) {
	parts := strings.Split(strings.Trim(path, "`/"), "/")
	if len(parts) != 3 || len(parts[0]) == 0 || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return streamQualifier{}, fmt.Errorf("Expected stream path with three components (i.e. 'organization/project/stream'), but got '%s'", path)
	}

	sq := streamQualifier{
		Organization: strings.ReplaceAll(parts[0], "-", "_"),
		Project:      strings.ReplaceAll(parts[1], "-", "_"),
		Stream:       strings.ReplaceAll(strings.TrimPrefix(parts[2], "stream:"), "-", "_"),
	}

	return sq, nil
}

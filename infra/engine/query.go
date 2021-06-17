package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

const (
	maxReferencedTables = 5
)

var (
	warehouseQueryTableRegex = regexp.MustCompile("`(/?[_\\-a-zA-Z0-9]+/[_:\\-/a-zA-Z0-9]+)`")
)

// WarehouseQueryTableResolver is a callback that should resolve a table qualifier
type WarehouseQueryTableResolver func(ctx context.Context, organization, project, table string) (driver.Project, driver.Table, driver.TableInstance, error)

// ExpandWarehouseQuery replaces table references of the form `org/proj/table` in warehouse queries
// with the proper name of the underlying table in the warehouse
func (e *Engine) ExpandWarehouseQuery(ctx context.Context, query string, resolver WarehouseQueryTableResolver) (string, error) {
	// find table paths in query
	matches := warehouseQueryTableRegex.FindAllString(query, -1)
	if len(matches) == 0 {
		return query, nil
	}

	// create mapping of tables to resolve
	mu := &sync.Mutex{}
	tableNames := make(map[tableQualifier]string)
	for _, path := range matches {
		qualifier, err := newTableQualifier(path)
		if err != nil {
			return "", err
		}
		tableNames[qualifier] = ""
	}

	// check number of tables is within limit
	if len(tableNames) > maxReferencedTables {
		return "", fmt.Errorf("a query cannot reference more than %d tables", maxReferencedTables)
	}

	// use an errgroup to concurrently lookup all the tables referenced in the query
	group, gctx := errgroup.WithContext(ctx)
	for qualifier := range tableNames {
		qualifier := qualifier // bind locally
		group.Go(func() error {
			p, s, si, err := resolver(gctx, qualifier.Organization, qualifier.Project, qualifier.Table)
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
	expandedQuery := warehouseQueryTableRegex.ReplaceAllStringFunc(query, func(path string) string {
		qualifier, err := newTableQualifier(path)
		if err != nil {
			panic(err) // shouldn't happen; would have been caught earlier
		}
		return tableNames[qualifier]
	})

	return expandedQuery, nil
}

type tableQualifier struct {
	Organization string
	Project      string
	Table        string
}

func newTableQualifier(path string) (tableQualifier, error) {
	parts := strings.Split(strings.Trim(path, "`/"), "/")
	if len(parts) != 3 || len(parts[0]) == 0 || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return tableQualifier{}, fmt.Errorf("Expected table path with three components (i.e. 'organization/project/table'), but got '%s'", path)
	}

	sq := tableQualifier{
		Organization: strings.ReplaceAll(parts[0], "-", "_"),
		Project:      strings.ReplaceAll(parts[1], "-", "_"),
		Table:        strings.ReplaceAll(strings.TrimPrefix(parts[2], "table:"), "-", "_"),
	}

	return sq, nil
}

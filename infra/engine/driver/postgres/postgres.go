package postgres

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

// Postgres implements LookupService and WarehouseService
type Postgres struct {
}

// Options for connecting to Postgres
type Options struct {
}

func init() {
	driver.AddDriver("postgres", newPostgres)
}

func newPostgres(optsMap map[string]interface{}) (driver.Service, error) {
	var opts Options
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding postgres options: %s", err.Error())
	}

	return &Postgres{}, nil
}

// AsLookupService implements Service
func (p *Postgres) AsLookupService() driver.LookupService {
	return p
}

// AsWarehouseService implements Service
func (p *Postgres) AsWarehouseService() driver.WarehouseService {
	return p
}

// AsUsageService implements Service
func (p *Postgres) AsUsageService() driver.UsageService {
	return p
}

package mock

import (
	"gitlab.com/beneath-hq/beneath/engine/driver"
)

// Mock implements beneath.WarehouseService
type Mock struct{}

// GetWarehouseService returns a Postgres implementation of beneath.LookupService
func GetWarehouseService() driver.WarehouseService {
	return Mock{}
}

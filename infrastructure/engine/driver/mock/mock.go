package mock

import (
	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
)

// Mock implements WarehouseService
type Mock struct {
}

func init() {
	driver.AddDriver("mock", newMock)
}

func newMock(optsMap map[string]interface{}) (driver.Service, error) {
	return &Mock{}, nil
}

// AsLookupService implements Service
func (p *Mock) AsLookupService() driver.LookupService {
	return nil
}

// AsWarehouseService implements Service
func (p *Mock) AsWarehouseService() driver.WarehouseService {
	return p
}

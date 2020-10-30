package engine

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/mathutil"
)

// Engine interfaces with the data layer
type Engine struct {
	Lookup    driver.LookupService
	Warehouse driver.WarehouseService
	Usage     driver.UsageService

	maxBatchLength int
	maxRecordSize  int
	maxKeySize     int
}

// IndexOptions configures the index service
type IndexOptions struct {
	DriverName    string                 `mapstructure:"driver"`
	DriverOptions map[string]interface{} `mapstructure:",remain"`
}

// WarehouseOptions configures the warehouse service
type WarehouseOptions struct {
	DriverName    string                 `mapstructure:"driver"`
	DriverOptions map[string]interface{} `mapstructure:",remain"`
}

// NewEngine constructs a new Engine for the given drivers
func NewEngine(indexOpts *IndexOptions, warehouseOpts *WarehouseOptions) (*Engine, error) {
	// make lookup service
	lookupServiceT, err := getService(indexOpts.DriverName, indexOpts.DriverOptions)
	if err != nil {
		return nil, err
	}
	lookupService := lookupServiceT.AsLookupService()
	if lookupService == nil {
		return nil, fmt.Errorf("driver '%s' is not a valid index service driver", indexOpts.DriverName)
	}

	// make warehouse service
	var warehouseService driver.WarehouseService
	if warehouseOpts.DriverName == indexOpts.DriverName {
		warehouseService = lookupServiceT.AsWarehouseService()
		if warehouseService == nil {
			return nil, fmt.Errorf("driver '%s' is not a valid warehouse service driver", warehouseOpts.DriverName)
		}
	} else {
		warehouseServiceT, err := getService(warehouseOpts.DriverName, warehouseOpts.DriverOptions)
		if err != nil {
			return nil, err
		}
		warehouseService = warehouseServiceT.AsWarehouseService()
		if warehouseService == nil {
			return nil, fmt.Errorf("driver '%s' is not a valid warehouse service driver", warehouseOpts.DriverName)
		}
	}

	// make usage service. THIS IS A HACK.
	usageService := lookupService.AsUsageService()
	if usageService == nil {
		return nil, fmt.Errorf("expected lookup service to double as a usage service")
	}

	e := &Engine{
		Lookup:         lookupService,
		Warehouse:      warehouseService,
		Usage:          usageService,
		maxBatchLength: mathutil.MinInts(lookupService.MaxRecordsInBatch(), warehouseService.MaxRecordsInBatch()),
		maxRecordSize:  mathutil.MinInts(lookupService.MaxRecordSize(), warehouseService.MaxRecordSize()),
		maxKeySize:     mathutil.MinInts(lookupService.MaxKeySize(), warehouseService.MaxKeySize()),
	}

	return e, nil
}

func getService(driverName string, driverOpts map[string]interface{}) (driver.Service, error) {
	driver, ok := driver.Drivers[driverName]
	if !ok {
		return nil, fmt.Errorf("no driver found with name '%s'", driverName)
	}

	return driver(driverOpts)
}

// Healthy returns true if connected to all services
func (e *Engine) Healthy() bool {
	return true
}

// CheckRecordSize validates that the record fits within the constraints of the underlying infrastructure
func (e *Engine) CheckRecordSize(s driver.Stream, structured map[string]interface{}, avroBytesLen int) error {
	if avroBytesLen > e.maxRecordSize {
		return fmt.Errorf("encoded record size exceeds maximum of %d bytes", e.maxRecordSize)
	}

	codec := s.GetCodec()

	err := e.checkKeySize(codec, codec.PrimaryIndex, structured)
	if err != nil {
		return err
	}

	for _, index := range codec.SecondaryIndexes {
		err := e.checkKeySize(codec, index, structured)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) checkKeySize(codec *codec.Codec, index codec.Index, structured map[string]interface{}) error {
	key, err := codec.MarshalKey(index, structured)
	if err != nil {
		return err
	}

	if len(key) > e.maxKeySize {
		return fmt.Errorf("encoded key size for index on fields %v exceeds maximum length of %d bytes", index.GetFields(), e.maxKeySize)
	}

	return nil
}

// CheckBatchLength validates that the number of records in a batch fits within the constraints of the underlying infrastructure
func (e *Engine) CheckBatchLength(length int) error {
	if length > e.maxBatchLength {
		return fmt.Errorf("batch length exceeds maximum of %d", e.maxBatchLength)
	}
	return nil
}

// Reset resets the state of the engine (useful during testing)
func (e *Engine) Reset(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Lookup.Reset(ctx)
	})

	group.Go(func() error {
		return e.Warehouse.Reset(ctx)
	})

	return group.Wait()
}

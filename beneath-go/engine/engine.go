package engine

import (
	"context"
	"fmt"

	"github.com/beneath-core/beneath-go/engine/driver"
	"github.com/beneath-core/beneath-go/engine/driver/bigquery"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable"
	"github.com/beneath-core/beneath-go/engine/driver/pubsub"
	pb "github.com/beneath-core/beneath-go/proto"
)

// Engine interfaces with the data layer
type Engine struct {
	MQ        driver.MessageQueue
	Log       driver.Log
	Lookup    driver.LookupService
	Warehouse driver.WarehouseService
}

// NewEngine creates a new Engine instance
func NewEngine(mqDriver, logDriver, lookupDriver, warehouseDriver string) *Engine {
	engine := &Engine{}
	engine.MQ = makeMQ(mqDriver)
	engine.Log = makeLog(logDriver)
	engine.Lookup = makeLookup(lookupDriver)
	engine.Warehouse = makeWarehouse(warehouseDriver)
	return engine
}

func makeMQ(driver string) driver.MessageQueue {
	switch driver {
	case "pubsub":
		return pubsub.GetMessageQueue()
	default:
		panic(fmt.Errorf("invalid mq driver '%s'", driver))
	}
}

func makeLog(driver string) driver.Log {
	switch driver {
	case "bigtable":
		return bigtable.GetLog()
	default:
		panic(fmt.Errorf("invalid log driver '%s'", driver))
	}
}

func makeLookup(driver string) driver.LookupService {
	switch driver {
	case "bigtable":
		return bigtable.GetLookupService()
	default:
		panic(fmt.Errorf("invalid lookup driver '%s'", driver))
	}
}

func makeWarehouse(driver string) driver.WarehouseService {
	switch driver {
	case "bigquery":
		return bigquery.GetWarehouseService()
	default:
		panic(fmt.Errorf("invalid warehouse driver '%s'", driver))
	}
}

// Healthy returns true if connected to all services
func (e *Engine) Healthy() bool {
	// TODO
	return true
}

// CheckRecordSize validates that the record fits within the constraints of the underlying infrastructure
func (e *Engine) CheckRecordSize(keyBytesLen int, avroBytesLen int) error {
	// TODO
	return nil
}

// CheckBatchLength validates that the number of records in a batch fits within the constraints of the underlying infrastructure
func (e *Engine) CheckBatchLength(length int) error {
	// TODO
	return nil
}

// QueueTask queues a task for processing
func (e *Engine) QueueTask(ctx context.Context, t *pb.QueuedTask) error {
	panic("todo")
}

// ReadTasks reads queued tasks
func (e *Engine) ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error {
	panic("todo")
}

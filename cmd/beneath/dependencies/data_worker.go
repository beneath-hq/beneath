package dependencies

import (
	"context"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/services/data"
)

func init() {
	// Like the control worker, the data worker is so lean it's defined below in this file.
	// It basically just calls RunWorker on a *data.Service.
	cli.AddDependency(NewDataWorker)
}

// DataWorker is the data background worker (that writes data to the engine)
type DataWorker struct {
	DataService *data.Service
}

// NewDataWorker initializes a new DataWorker
func NewDataWorker(bus *bus.Bus, data *data.Service) *DataWorker {
	return &DataWorker{
		DataService: data,
	}
}

// Run starts the worker
func (s *DataWorker) Run(ctx context.Context) error {
	return s.DataService.RunWorker(ctx)
}

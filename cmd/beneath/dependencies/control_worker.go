package dependencies

import (
	"context"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
)

func init() {
	// The control worker is so lean, it's defined below in this file.
	// It basically just ensures all services are created in the process (to register
	// their events handlers with the bus, and then calls bus.Run)

	cli.AddDependency(NewControlWorker)

	cli.AddStartable(&cli.Startable{
		Name: "control-worker",
		Register: func(lc *cli.Lifecycle, worker *ControlWorker) {
			lc.Add("control-worker", worker)
		},
	})

}

// ControlWorker is the background worker for async events (i.e. it runs bus.Run)
type ControlWorker struct {
	Bus      *bus.Bus
	Services *AllServices // Ensures all services are created and registers event handlers with Bus
}

// NewControlWorker initializes a new ControlWorker
func NewControlWorker(bus *bus.Bus, services *AllServices) *ControlWorker {
	return &ControlWorker{
		Bus:      bus,
		Services: services,
	}
}

// Run starts the worker
func (s *ControlWorker) Run(ctx context.Context) error {
	return s.Bus.Run(ctx)
}

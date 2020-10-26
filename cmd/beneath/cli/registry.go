package cli

import (
	"context"
	"os"

	"gitlab.com/beneath-hq/beneath/pkg/ctxutil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"golang.org/x/sync/errgroup"
)

// the registered dependencies
var dependencies []interface{}

// AddDependency registers a dependency to provide to fx.
// A dependency is a constructor, i.e. a "New..." function
func AddDependency(constructor interface{}) {
	dependencies = append(dependencies, constructor)
}

// Startable is a service that can be started with the start command
type Startable struct {
	Name     string
	Register interface{} // func (lc *Lifecycle, s *Startable) { lc.Add(s.Run) }
}

// the registered startables
var startables []*Startable

// AddStartable registers a startable
func AddStartable(s *Startable) {
	startables = append(startables, s)
}

// ConfigKey represents a config key and optionally a default value
type ConfigKey struct {
	Key         string
	Description string
	Default     interface{}
}

// the registered config keys
var configKeys []*ConfigKey

// AddConfigKey registers a new config key to be loaded from config files
// with Viper and/or overriden via the CLI
func AddConfigKey(key *ConfigKey) {
	configKeys = append(configKeys, key)
}

// Runable should implement a Run function that returns nil on success,
// an error on failure, and stops gracefully if the ctx is cancelled.
type Runable interface {
	Run(ctx context.Context) error
}

// Lifecycle runs one or more Runables. It is inspired by fx.Lifecycle,
// but we prefer a ctx-based Run function to start/stop hooks.
type Lifecycle struct {
	Runners []Runable
}

// Add registers a new Runable in the lifecycle
func (l *Lifecycle) Add(runner Runable) {
	l.Runners = append(l.Runners, runner)
}

// Run is where the magic happens! It runs every Runner registered in the lifecycle.
// If any runner returns an error, it cancels every other runner and stops the application.
// It also cancels runners if it receives a termination signal (ie. SIGINT or SIGTERM).
func (l *Lifecycle) Run() {
	// A ctx that we can cancel, and that also cancels on sigint, etc.
	ctx, cancel := context.WithCancel(ctxutil.WithCancelOnTerminate(context.Background()))

	// Run every runner in an error group
	group := new(errgroup.Group)
	for _, runner := range l.Runners {
		runner := runner
		group.Go(func() error {
			log.S.Infof("running %T", runner)
			err := runner.Run(ctx)
			log.S.Infof("%T stopped", runner)
			if err != nil {
				cancel()
			}
			return err
		})
	}

	err := group.Wait()
	if err != nil {
		log.S.Fatal(err)
	}
	os.Exit(0)
}

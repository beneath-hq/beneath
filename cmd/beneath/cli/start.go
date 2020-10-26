package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"

	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// registers start command, which runs services registered with AddStartable
func (c *CLI) newStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start backend service",
		Long:  `Starts backend servers or workers`,
	}

	for _, service := range startables {
		serviceCopy := service
		cmd := &cobra.Command{
			Use:  service.Name,
			Args: cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				c.runStartableServices([]*Startable{serviceCopy})
			},
		}
		startCmd.AddCommand(cmd)
	}

	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Starts every backend service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c.runStartableServices(startables)
		},
	}
	startCmd.AddCommand(allCmd)

	return startCmd
}

// runs the given startable services (with dependencies injected through fx)
func (c *CLI) runStartableServices(services []*Startable) {
	log.InitLogger()

	var invokers []interface{}
	for _, service := range services {
		invokers = append(invokers, service.Register)
	}

	var lc *Lifecycle
	fx.New(
		fx.NopLogger,
		fx.Provide(func() *Lifecycle { return &Lifecycle{} }), // provides Lifecycle object (see registry.go for more)
		fx.Provide(func() *viper.Viper { return c.v }),        // provides viper config for other dependencies to consume
		fx.Provide(dependencies...),                           // all dependencies registered with AddDependency
		fx.Invoke(invokers...),                                // registers every startable
		fx.Populate(&lc),                                      // extract lifecycle
	)

	// migrations.MustRunUp(hub.DB)
	lc.Run()
}

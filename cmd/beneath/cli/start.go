package cli

import (
	"github.com/spf13/cobra"

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

	// At this point, every dependency should already have been registered with AddDependency

	// Provides the Lifecycle object (see registry.go for more)
	AddDependency(func() *Lifecycle { return &Lifecycle{} })

	// Invoke every service's register function
	for _, service := range services {
		err := Dig.Invoke(service.Register)
		if err != nil {
			panic(err)
		}
	}

	// Invoke lifecycle
	Dig.Invoke(func(lc *Lifecycle) {
		lc.Run()
	})
}

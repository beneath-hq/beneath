package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *CLI) newRootCmd() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "beneath",
		Short: "Open source data systems meta-layer",
		Long:  "Beneath is the management layer for all your data systems.\nSee https://beneath.dev for details.",
		// Called before every command AND subcommand.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Loads config file into c.v
			err := c.loadConfig(configFile)
			if err != nil {
				return err
			}

			// Provides viper config for other dependencies to consume
			AddDependency(func() *viper.Viper { return c.v })
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./config/.default.yaml merged with ./config/.ENV.yaml)")
	return cmd
}

package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CLI builds and serves the CLI for *launching* the Beneath backend (*not* the Python-based CLI
// for local authentication, creating streams, etc.)
type CLI struct {
	rootCmd *cobra.Command
	v       *viper.Viper
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	c := &CLI{}
	c.v = viper.New()
	c.rootCmd = c.newRootCmd()
	return c
}

// Run parses and runs commands
func (c *CLI) Run() {
	// add subcommands
	c.rootCmd.AddCommand(c.newStartCmd())

	// run
	if err := c.rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func (c *CLI) newRootCmd() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "beneath",
		Short: "Open source data systems meta-layer",
		Long:  "Beneath is the management layer for all your data systems.\nSee https://beneath.dev for details.",
		// Called before every command AND subcommand.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Loads config file into c.v
			return c.loadConfig(configFile)
		},
	}
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./config/.env)")
	return cmd
}

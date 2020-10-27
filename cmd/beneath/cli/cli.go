package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CLI builds and serves the CLI for *launching* the Beneath backend (*not* the Python-based CLI
// for local authentication, creating streams, etc.)
type CLI struct {
	Root *cobra.Command
	v    *viper.Viper
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	c := &CLI{}
	c.v = viper.New()
	c.Root = c.newRootCmd()
	c.Root.AddCommand(c.newStartCmd())
	return c
}

// Run parses and runs commands
func (c *CLI) Run() {
	if err := c.Root.Execute(); err != nil {
		os.Exit(1)
	}
}

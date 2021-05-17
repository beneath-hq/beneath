package dependencies

import (
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/server/control"
	"github.com/spf13/viper"
)

func init() {
	cli.AddDependency(control.NewServer)

	cli.AddDependency(func(v *viper.Viper) (*control.ServerOptions, error) {
		var opts control.ServerOptions
		return &opts, v.UnmarshalKey("control", &opts)
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.port",
		Default:     "4000",
		Description: "control server port",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.host",
		Default:     "http://localhost:4000",
		Description: "control server host (with protocol) (for redirects)",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.frontend_host",
		Default:     "http://localhost:3000",
		Description: "host (with protocol) of the web frontend (for redirects)",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.session_secret",
		Default:     "",
		Description: "secret to use for signing session tokens",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.auth.github.id",
		Default:     "",
		Description: "id for authentication with Github",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.auth.github.secret",
		Default:     "",
		Description: "secret for authentication with Github",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.auth.google.id",
		Default:     "",
		Description: "id for authentication with Google",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.auth.google.secret",
		Default:     "",
		Description: "secret for authentication with Google",
	})
}

package dependencies

import (
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/server/data"
	"github.com/spf13/viper"
)

func init() {
	cli.AddDependency(data.NewServer)

	cli.AddDependency(func(v *viper.Viper) (*data.ServerOptions, error) {
		var opts data.ServerOptions
		return &opts, v.UnmarshalKey("data", &opts)
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "data.http_port",
		Default:     "5000",
		Description: "data server port for HTTP",
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "data.grpc_port",
		Default:     "50051",
		Description: "data server port for GRPC",
	})
}

package dependencies

import (
	"github.com/spf13/viper"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/server/data"
	"gitlab.com/beneath-hq/beneath/services/metrics"
)

func init() {
	cli.AddStartable(&cli.Startable{
		Name: "data-server",
		Register: func(lc *cli.Lifecycle, server *data.Server, metrics *metrics.Broker) {
			lc.Add(server)
			lc.Add(metrics)
		},
	})

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

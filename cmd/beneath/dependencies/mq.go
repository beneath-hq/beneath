package dependencies

import (
	"github.com/spf13/viper"
	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/infrastructure/mq"

	// registers all mq drivers
	_ "gitlab.com/beneath-hq/beneath/infrastructure/mq/driver/pubsub"
)

func init() {
	cli.AddDependency(mq.NewMessageQueue)
	cli.AddDependency(func(v *viper.Viper) (*mq.Options, error) {
		var opts mq.Options
		return &opts, v.UnmarshalKey("mq", &opts)
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "mq.driver",
		Default:     "",
		Description: "driver to use for message queue",
	})
}

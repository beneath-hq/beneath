package dependencies

import (
	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/infra/mq"
	"github.com/spf13/viper"

	// registers all mq drivers
	_ "github.com/beneath-hq/beneath/infra/mq/driver/pubsub"
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
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "mq.subscriber_id",
		Default:     "",
		Description: "unique identifier for the subscriber",
	})
}

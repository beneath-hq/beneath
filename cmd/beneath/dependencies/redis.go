package dependencies

import (
	"github.com/spf13/viper"

	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/infra/redis"
)

func init() {
	cli.AddDependency(redis.NewRedis)
	cli.AddDependency(func(v *viper.Viper) (*redis.Options, error) {
		var opts redis.Options
		return &opts, v.UnmarshalKey("control.redis", &opts)
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.redis.url",
		Default:     "redis://localhost/",
		Description: "Redis connection URL for control-plane caching",
	})

}

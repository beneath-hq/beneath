package dependencies

import (
	"github.com/spf13/viper"

	"github.com/beneath-hq/beneath/cmd/beneath/cli"
	"github.com/beneath-hq/beneath/infra/db"
)

func init() {
	cli.AddDependency(db.NewDB)
	cli.AddDependency(func(v *viper.Viper) (*db.Options, error) {
		var opts db.Options
		return &opts, v.UnmarshalKey("control.postgres", &opts)
	})

	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.postgres.host",
		Default:     "localhost",
		Description: "Postgres host for control-plane db",
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.postgres.database",
		Default:     "postgres",
		Description: "Postgres database for control-plane db",
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.postgres.user",
		Default:     "postgres",
		Description: "Postgres user for control-plane db",
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "control.postgres.password",
		Default:     "",
		Description: "Postgres password for control-plane db",
	})
}

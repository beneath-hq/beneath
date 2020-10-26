package dependencies

import (
	"github.com/spf13/viper"

	"gitlab.com/beneath-hq/beneath/cmd/beneath/cli"
	"gitlab.com/beneath-hq/beneath/infrastructure/engine"

	// registers all engine drivers
	_ "gitlab.com/beneath-hq/beneath/infrastructure/engine/driver/bigquery"
	_ "gitlab.com/beneath-hq/beneath/infrastructure/engine/driver/bigtable"
	_ "gitlab.com/beneath-hq/beneath/infrastructure/engine/driver/mock"
	_ "gitlab.com/beneath-hq/beneath/infrastructure/engine/driver/postgres"
)

func init() {
	cli.AddDependency(engine.NewEngine)
	cli.AddDependency(func(v *viper.Viper) (*engine.IndexOptions, error) {
		var indexOpts engine.IndexOptions
		err := v.UnmarshalKey("data.index", &indexOpts)
		if err != nil {
			return nil, err
		}
		return &indexOpts, nil
	})
	cli.AddDependency(func(v *viper.Viper) (*engine.WarehouseOptions, error) {
		var warehouseOpts engine.WarehouseOptions
		err := v.UnmarshalKey("data.warehouse", &warehouseOpts)
		if err != nil {
			return nil, err
		}
		return &warehouseOpts, nil
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "data.index.driver",
		Default:     "",
		Description: "driver to use for (indexed) operational serving of stream records",
	})
	cli.AddConfigKey(&cli.ConfigKey{
		Key:         "data.warehouse.driver",
		Default:     "",
		Description: "driver to use for OLAP queries of stream records",
	})
	// 	Key:     "data.log.driver",
	// 	Default: "",
	// 	Description:   "driver to use for log storage of stream records",
}

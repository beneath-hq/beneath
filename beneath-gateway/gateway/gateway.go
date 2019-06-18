package gateway

import (
	"github.com/beneath-core/beneath-gateway/beneath/core"
	"github.com/beneath-core/beneath-gateway/beneath/engine"
	"github.com/beneath-core/beneath-gateway/beneath/management"
)

type configSpecification struct {
	HTTPPort        int    `envconfig:"HTTP_PORT" default:"5000"`
	GRPCPort        int    `envconfig:"GRPC_PORT" default:"50051"`
	StreamsDriver   string `envconfig:"STREAMS_DRIVER" required:"true"`
	TablesDriver    string `envconfig:"TABLES_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"WAREHOUSE_DRIVER" required:"true"`
	RedisURL        string `envconfig:"REDIS_URL" required:"true"`
	PostgresURL     string `envconfig:"POSTGRES_URL" required:"true"`
}

var (
	// Config for gateway
	Config configSpecification

	// Engine is the data plane
	Engine *engine.Engine
)

func init() {
	core.LoadConfig("beneath", &Config)
	Engine = engine.NewEngine(Config.StreamsDriver, Config.TablesDriver)
	management.Init(Config.PostgresURL, Config.RedisURL)
}

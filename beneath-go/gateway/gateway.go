package gateway

import (
	"github.com/beneath-core/beneath-go/control/db"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine"
)

type configSpecification struct {
	HTTPPort        int    `envconfig:"GATEWAY_PORT" default:"5000"`
	GRPCPort        int    `envconfig:"GATEWAY_PORT_GRPC" default:"50051"`
	WSPort          int    `envconfig:"GATEWAY_PORT_WS" default:"8080"`
	StreamsDriver   string `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
	TablesDriver    string `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL        string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresURL     string `envconfig:"CONTROL_POSTGRES_URL" required:"true"`
}

var (
	// Config for gateway
	Config configSpecification

	// Engine is the data plane
	Engine *engine.Engine
)

func init() {
	core.LoadConfig("beneath", &Config)

	Engine = engine.NewEngine(Config.StreamsDriver, Config.TablesDriver, Config.WarehouseDriver)

	db.InitPostgres(Config.PostgresURL)
	db.InitRedis(Config.RedisURL)
}

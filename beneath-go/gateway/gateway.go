package gateway

import (
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine"
	"github.com/beneath-core/beneath-go/management"
)

type configSpecification struct {
	HTTPPort        int    `envconfig:"GATEWAY_PORT" default:"5000"`
	GRPCPort        int    `envconfig:"GATEWAY_PORT_GRPC" default:"50051"`
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

	// InstanceCache to find instanceIDs when user requests by project/stream name
	InstanceCache management.InstanceCache

	// StreamCache to find schemas and other details for an instanceID
	StreamCache management.StreamCache

	// RoleCache to find permissions of a key relative to a projectID
	RoleCache management.RoleCache
)

func init() {
	core.LoadConfig("beneath", &Config)
	Engine = engine.NewEngine(Config.StreamsDriver, Config.TablesDriver)
	management.Init(Config.PostgresURL, Config.RedisURL)
	InstanceCache = management.NewInstanceCache()
	StreamCache = management.NewStreamCache()
	RoleCache = management.NewRoleCache()
}

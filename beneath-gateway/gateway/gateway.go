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

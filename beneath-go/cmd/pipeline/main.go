package main

import (
	"context"
	"fmt"
	"time"
	
	uuid "github.com/satori/go.uuid"
	
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/pipeline"
	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

func main() {
	config configSpecification
	core.LoadConfig("beneath", &config)

	db.InitPostgres(Config.PostgresHost, Config.PostgresUser, Config.PostgresPassword)
	db.InitRedis(Config.RedisURL)
	db.InitEngine(Config.MQDriver, Config.LookupDriver, Config.WarehouseDriver)

	log.S.Info("pipeline started")
	log.S.Fatal(pipeline.Run())
}

package main

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/taskqueue"
)

type configSpecification struct {
	StreamsDriver   string `envconfig:"ENGINE_STREAMS_DRIVER" required:"true"`
	TablesDriver    string `envconfig:"ENGINE_TABLES_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
}

func main() {
	var config configSpecification
	core.LoadConfig("beneath", &config)
	db.InitEngine(config.StreamsDriver, config.TablesDriver, config.WarehouseDriver)

	err := taskqueue.Submit(context.Background(), &entity.RunBillingTask{})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	log.S.Info("Successfully scheduled task")
}

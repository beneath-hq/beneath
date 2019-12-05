package main

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/taskqueue"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
)

type configSpecification struct {
	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LogDriver       string `envconfig:"ENGINE_LOG_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
}

func main() {
	var config configSpecification
	core.LoadConfig("beneath", &config)
	db.InitEngine(Config.MQDriver, Config.LogDriver, Config.LookupDriver, Config.WarehouseDriver)

	err := taskqueue.Submit(context.Background(), &entity.RunBillingTask{})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	log.S.Info("Successfully scheduled task")
}

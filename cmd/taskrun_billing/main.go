package main

import (
	"context"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/envutil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

type configSpecification struct {
	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
}

func main() {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)
	log.InitLogger()
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	err := taskqueue.Submit(context.Background(), &entity.RunBillingTask{})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	log.S.Info("Successfully scheduled task")
}

package main

import (
	"context"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/taskqueue"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/internal/hub"
)

type configSpecification struct {
	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
}

func main() {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)

	err := taskqueue.Submit(context.Background(), &entity.RunBillingTask{})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	log.S.Info("Successfully scheduled task")
}

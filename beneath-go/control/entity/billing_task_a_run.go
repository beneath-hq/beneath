package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/control/taskqueue"
	"github.com/beneath-core/beneath-go/core/log"
)

// RunBillingTask triggers billing (computation, invoice creation, and stripe)
type RunBillingTask struct {
}

// register task
func init() {
	taskqueue.RegisterTask(&RunBillingTask{})
}

// Run triggers the task
// it is expected that the task is run at the beginning of each month
// organizations will be assessed usage and corresponding overage fees for the previous month
// organizations will be charged seats for the upcoming month
func (t *RunBillingTask) Run(ctx context.Context) error {
	organizations := FindAllOrganizations(ctx)
	timestamp := time.Now()

	for _, o := range organizations {
		err := taskqueue.Submit(context.Background(), &ComputeBillResourcesTask{
			OrganizationID: o.OrganizationID,
			Timestamp:      timestamp,
		})
		if err != nil {
			log.S.Errorw("Error creating task", err)
		}
	}

	return nil
}
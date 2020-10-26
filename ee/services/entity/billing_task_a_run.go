package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// RunBillingTask triggers billing (computation, invoice creation, and stripe)
type RunBillingTask struct {
}

// Run triggers the task
// it is expected that the task is run at the beginning of each month
// organizations will be assessed usage and corresponding overage fees for the previous month
// organizations will be charged seats and prepaid quotas for the upcoming month
func (t *RunBillingTask) Run(ctx context.Context) error {
	billingInfos := FindAllPayingBillingInfos(ctx)
	timestamp := time.Now()
	timestamp = timestamp.AddDate(0, 1, 0) // FOR TESTING: simulate that we are in the next month

	for _, bi := range billingInfos {
		err := hub.Bus.Publish(ctx, &ComputeBillResourcesTask{
			BillingInfo: bi,
			Timestamp:   timestamp,
		})
		if err != nil {
			log.S.Errorw("Error creating task", err)
		}
	}

	return nil
}

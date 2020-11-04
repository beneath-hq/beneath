package billing

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// RunBilling runs billing for every customer with next_billing_time <  now
func (s *Service) RunBilling(ctx context.Context) error {
	log.S.Infof("Running billing at %s", time.Now().String())

	bis := s.FindBillingInfosToRun(ctx)
	for idx, bi := range bis {
		log.S.Infow(fmt.Sprintf("Running billing %d/%d", idx, len(bis)), "organization_id", bi.OrganizationID.String())
		err := s.CommitBilledResources(ctx, bi)
		if err != nil {
			return err
		}
	}

	log.S.Infof("Completed billing for %d customers", len(bis))
	return nil
}

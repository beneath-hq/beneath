package billing

import (
	"context"
	"fmt"
)

// RunBilling runs billing for every customer with next_billing_time <  now
func (s *Service) RunBilling(ctx context.Context) error {
	s.Logger.Infof("running billing")

	bis := s.FindBillingInfosToRun(ctx)
	for idx, bi := range bis {
		s.Logger.Infow(fmt.Sprintf("running billing %d/%d", idx, len(bis)), "organization_id", bi.OrganizationID.String())
		err := s.CommitBilledResources(ctx, bi)
		if err != nil {
			return err
		}
	}

	s.Logger.Infof("completed billing for %d customers", len(bis))
	return nil
}

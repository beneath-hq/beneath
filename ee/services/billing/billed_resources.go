package billing

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/ee/models"
)

// FindBilledResources returns the matching billed resources or nil
func (s *Service) FindBilledResources(ctx context.Context, organizationID uuid.UUID, billingTime time.Time) []*models.BilledResource {
	var billedResources []*models.BilledResource
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time = ?", billingTime).
		Select()
	if err != nil {
		panic(err)
	}
	return billedResources
}

// CreateOrUpdateBilledResources upserts billed resources
func (s *Service) CreateOrUpdateBilledResources(ctx context.Context, billedResources []*models.BilledResource) error {
	// upsert
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, &billedResources).
		OnConflict("(billing_time, organization_id, entity_id, product) DO UPDATE").
		Set("start_time = EXCLUDED.start_time").
		Set("end_time = EXCLUDED.end_time").
		Set("quantity = EXCLUDED.quantity").
		Set("total_price_cents = EXCLUDED.total_price_cents").
		Set("currency = EXCLUDED.currency").
		Set("updated_on = now()").
		Insert()
	if err != nil {
		return err
	}

	return nil
}

package billing

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/ee/models"
	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	nee_models "gitlab.com/beneath-hq/beneath/models"
)

// FindBilledResources returns the matching billed resources or nil
func (s *Service) FindBilledResources(ctx context.Context, organizationID uuid.UUID, fromTime, toTime time.Time) []*models.BilledResource {
	var billedResources []*models.BilledResource
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time >= ?", fromTime).
		Where("billing_time < ?", toTime).
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

// CommitBaseBilledResource upserts a billed resource for the base price of the billing plan
func (s *Service) CommitBaseBilledResource(ctx context.Context, bi *models.BillingInfo, billingTime, startTime, endTime time.Time) error {
	if bi.BillingPlan.BasePriceCents == 0 {
		return nil
	}

	resource := &models.BilledResource{
		OrganizationID:  bi.OrganizationID,
		BillingTime:     billingTime,
		EntityID:        bi.OrganizationID,
		EntityKind:      models.OrganizationEntityKind,
		StartTime:       startTime,
		EndTime:         endTime,
		Product:         models.BaseProduct,
		Quantity:        1,
		TotalPriceCents: bi.BillingPlan.BasePriceCents,
		Currency:        bi.BillingPlan.Currency,
	}

	return s.CreateOrUpdateBilledResources(ctx, []*models.BilledResource{resource})
}

// CommitSeatBilledResources upserts billed resources for the seat price of every user in an organization
func (s *Service) CommitSeatBilledResources(ctx context.Context, bi *models.BillingInfo, members []*nee_models.OrganizationMember, billingTime, startTime, endTime time.Time) error {
	if bi.BillingPlan.SeatPriceCents == 0 {
		return nil
	}
	if len(members) == 0 {
		return nil
	}

	resources := make([]*models.BilledResource, len(members))
	for idx, member := range members {
		resources[idx] = &models.BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        member.UserID,
			EntityKind:      models.UserEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         models.SeatProduct,
			Quantity:        1,
			TotalPriceCents: bi.BillingPlan.SeatPriceCents,
			Currency:        bi.BillingPlan.Currency,
		}
	}

	return s.CreateOrUpdateBilledResources(ctx, resources)
}

// CommitProratedSeatBilledResource upserts a billed resource for a seat added mid-period
func (s *Service) CommitProratedSeatBilledResource(ctx context.Context, bi *models.BillingInfo, user *nee_models.User, billingTime, startTime, endTime time.Time, rate float64) error {
	if rate < 0 || rate >= 1 {
		return fmt.Errorf("cannot commit prorated seat with rate %f", rate)
	}

	price := int64(float64(bi.BillingPlan.SeatPriceCents) * rate)
	if price == 0 {
		return nil
	}

	resource := &models.BilledResource{
		OrganizationID:  bi.OrganizationID,
		BillingTime:     billingTime,
		EntityID:        user.UserID,
		EntityKind:      models.UserEntityKind,
		StartTime:       startTime,
		EndTime:         endTime,
		Product:         models.ProratedSeatProduct,
		Quantity:        rate,
		TotalPriceCents: price,
		Currency:        bi.BillingPlan.Currency,
	}

	return s.CreateOrUpdateBilledResources(ctx, []*models.BilledResource{resource})
}

// CommitOverageBilledResources upserts billed resources for the overage consumed
func (s *Service) CommitOverageBilledResources(ctx context.Context, bi *models.BillingInfo, usage pb.QuotaUsage, billingTime, startTime, endTime time.Time) error {
	readOverageResource := s.computeOverageResource(bi, billingTime, startTime, endTime, models.ReadOverageProduct, usage.ReadBytes, bi.Organization.PrepaidReadQuota, bi.BillingPlan.ReadOveragePriceCents)
	writeOverageResource := s.computeOverageResource(bi, billingTime, startTime, endTime, models.WriteOverageProduct, usage.WriteBytes, bi.Organization.PrepaidWriteQuota, bi.BillingPlan.WriteOveragePriceCents)
	scanOverageResource := s.computeOverageResource(bi, billingTime, startTime, endTime, models.ScanOverageProduct, usage.ScanBytes, bi.Organization.PrepaidScanQuota, bi.BillingPlan.ScanOveragePriceCents)

	var resources []*models.BilledResource
	if readOverageResource != nil {
		resources = append(resources, readOverageResource)
	}
	if writeOverageResource != nil {
		resources = append(resources, writeOverageResource)
	}
	if scanOverageResource != nil {
		resources = append(resources, scanOverageResource)
	}
	if len(resources) == 0 {
		return nil
	}

	return s.CreateOrUpdateBilledResources(ctx, resources)
}

func (s *Service) computeOverageResource(bi *models.BillingInfo, billingTime, startTime, endTime time.Time, product models.Product, bytes int64, prepaid *int64, priceCents int64) *models.BilledResource {
	if priceCents == 0 {
		return nil
	}

	overageBytes := bytes - derefInt64(prepaid, 0)
	overageGB := float64(overageBytes) / float64(1e9)
	overagePrice := int64(overageGB * float64(priceCents)) // NOTE: Overage prices are per GB

	if overageBytes <= 0 {
		return nil
	}

	return &models.BilledResource{
		OrganizationID:  bi.OrganizationID,
		BillingTime:     billingTime,
		EntityID:        bi.OrganizationID,
		EntityKind:      models.OrganizationEntityKind,
		StartTime:       startTime,
		EndTime:         endTime,
		Product:         product,
		Quantity:        overageGB,
		TotalPriceCents: overagePrice,
		Currency:        bi.BillingPlan.Currency,
	}
}

func derefInt64(val *int64, fallback int64) int64 {
	if val == nil {
		return fallback
	}
	return *val
}

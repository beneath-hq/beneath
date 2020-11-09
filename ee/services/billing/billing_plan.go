package billing

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/ee/models"
	"gitlab.com/beneath-hq/beneath/infra/db"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// FindBillingPlan finds a billing plan
func (s *Service) FindBillingPlan(ctx context.Context, billingPlanID uuid.UUID) *models.BillingPlan {
	billingPlan := &models.BillingPlan{
		BillingPlanID: billingPlanID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, billingPlan).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return billingPlan
}

// FindBillingPlansAvailableInUI finds the billing plans that should be available in the UI
func (s *Service) FindBillingPlansAvailableInUI(ctx context.Context) []*models.BillingPlan {
	var billingPlans []*models.BillingPlan
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billingPlans).
		Column("billing_plan.*").
		Where("available_in_ui = true").
		Select()
	if err != nil {
		panic(err)
	}
	return billingPlans
}

// GetDefaultBillingPlan returns the default billing plan, and caches it for future calls
func (s *Service) GetDefaultBillingPlan(ctx context.Context) *models.BillingPlan {
	return s.defaultBillingPlan.Get(ctx).(*models.BillingPlan)
}

// FindDefaultBillingPlan returns the current default billing plan
func (s *Service) FindDefaultBillingPlan(ctx context.Context) *models.BillingPlan {
	// find default plan
	plan := &models.BillingPlan{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, plan).Where(`"default" = true`).Select()
	if db.AssertFoundOne(err) {
		return plan
	}

	// create a default plan (OnConflict works because there's a unique index on default = true)
	plan = makeDefaultBillingPlan()
	_, err = s.DB.GetDB(ctx).Model(plan).OnConflict("DO NOTHING").Insert()
	if err != nil {
		panic(err)
	}

	// we'll refetch (in case of conflict)
	return s.FindDefaultBillingPlan(ctx)
}

func makeDefaultBillingPlan() *models.BillingPlan {
	return &models.BillingPlan{
		Default:        true,
		Description:    "Free plan",
		Currency:       models.DollarCurrency,
		Period:         timeutil.PeriodMonth,
		BaseReadQuota:  2000000000,
		BaseWriteQuota: 1000000000,
		BaseScanQuota:  100000000000,
		ReadQuota:      2000000000,
		WriteQuota:     1000000000,
		ScanQuota:      100000000000,
		MultipleUsers:  false,
		AvailableInUI:  true,
	}
}

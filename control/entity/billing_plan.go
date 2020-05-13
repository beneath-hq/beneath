package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// BillingPlan represents a Billing Plan that an Organization can subscribe to.
// Here's how the quota arithmetic works:
// prepaid quota = base quota + (seats * seat quota)
// potential overage = quota - prepaid quota
type BillingPlan struct {
	BillingPlanID          uuid.UUID       `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Default                bool            `sql:",notnull,default:false"`
	Description            string          `validate:"omitempty,lte=255"`
	CreatedOn              time.Time       `sql:",notnull,default:now()"`
	UpdatedOn              time.Time       `sql:",notnull,default:now()"`
	Currency               Currency        `sql:",notnull"`
	Period                 timeutil.Period `sql:",notnull"`
	BasePriceCents         int32           `sql:",notnull"`
	SeatPriceCents         int32           `sql:",notnull"`
	BaseReadQuota          int64           `sql:",notnull"` // bytes
	BaseWriteQuota         int64           `sql:",notnull"` // bytes
	SeatReadQuota          int64           `sql:",notnull"` // bytes
	SeatWriteQuota         int64           `sql:",notnull"` // bytes
	ReadQuota              int64           `sql:",notnull"` // bytes
	WriteQuota             int64           `sql:",notnull"` // bytes
	ReadOveragePriceCents  int32           `sql:",notnull"` // price per GB overage
	WriteOveragePriceCents int32           `sql:",notnull"` // price per GB overage
	MultipleUsers          bool            `sql:",notnull"`
	PrivateProjects        bool            `sql:",notnull"`
	AvailableInUI          bool            `sql:",notnull,default:false"`
}

// FindBillingPlan finds a billing plan by ID
func FindBillingPlan(ctx context.Context, billingPlanID uuid.UUID) *BillingPlan {
	billingPlan := &BillingPlan{
		BillingPlanID: billingPlanID,
	}
	err := hub.DB.ModelContext(ctx, billingPlan).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingPlan
}

// FindDefaultBillingPlan returns the current default billing plan
func FindDefaultBillingPlan(ctx context.Context) *BillingPlan {
	// find default plan
	plan := &BillingPlan{}
	err := hub.DB.ModelContext(ctx, plan).Where(`"default" = true`).Select()
	// if none was found
	if !AssertFoundOne(err) {
		// create a default plan (OnConflict works because there's a unique index on default = true)
		plan = makeDefaultBillingPlan()
		_, err := hub.DB.Model(plan).OnConflict("DO NOTHING").Insert()
		if err != nil {
			panic(err)
		}

		// we'll refetch (in case of conflict)
		err = hub.DB.Model(plan).Where(`"default" = true`).Select()
		if err != nil {
			panic(err)
		}
	}

	return plan
}

// FindBillingPlansAvailableInUI finds the billing plans available in the UI
func FindBillingPlansAvailableInUI(ctx context.Context) []*BillingPlan {
	var billingPlans []*BillingPlan
	err := hub.DB.ModelContext(ctx, &billingPlans).
		Column("billing_plan.*").
		Where("available_in_ui = true").
		Select()
	if err != nil {
		panic(err)
	}
	return billingPlans
}

func makeDefaultBillingPlan() *BillingPlan {
	return &BillingPlan{
		Default:         true,
		Description:     "Free",
		Currency:        DollarCurrency,
		Period:          timeutil.PeriodMonth,
		BaseReadQuota:   2000000000,
		BaseWriteQuota:  1000000000,
		ReadQuota:       2000000000,
		WriteQuota:      1000000000,
		MultipleUsers:   false,
		PrivateProjects: false,
		AvailableInUI:   true,
	}
}

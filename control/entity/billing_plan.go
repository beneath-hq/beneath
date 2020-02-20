package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/pkg/timeutil"
)

// BillingPlan represents a Billing Plan that an Organization can subscribe to
type BillingPlan struct {
	BillingPlanID          uuid.UUID       `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Default                bool            `sql:",notnull,default:false"`
	Description            string          `validate:"omitempty,lte=255"`
	CreatedOn              time.Time       `sql:",notnull,default:now()"`
	UpdatedOn              time.Time       `sql:",notnull,default:now()"`
	Currency               Currency        `sql:",notnull"`
	Period                 timeutil.Period `sql:",notnull"`
	SeatPriceCents         int32           `sql:",notnull"`
	SeatReadQuota          int64           `sql:",notnull"` // bytes
	SeatWriteQuota         int64           `sql:",notnull"` // bytes
	BaseReadQuota          int64           `sql:",notnull"` // bytes
	BaseWriteQuota         int64           `sql:",notnull"` // bytes
	ReadOveragePriceCents  int32           `sql:",notnull"` // price per GB overage
	WriteOveragePriceCents int32           `sql:",notnull"` // price per GB overage
	Personal               bool            `sql:",notnull"` // probably want to rename to "MultipleUsers" and flip the sign
	PrivateProjects        bool            `sql:",notnull"`
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
	var plan *BillingPlan
	err := hub.DB.Model(plan).Where(`"default" = true`).Select()

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

func makeDefaultBillingPlan() *BillingPlan {
	return &BillingPlan{
		Default:         true,
		Description:     "Default Billing Plan",
		Currency:        DollarCurrency,
		Period:          timeutil.PeriodMonth,
		SeatReadQuota:   10000000000,
		SeatWriteQuota:  10000000000,
		Personal:        true,
		PrivateProjects: false,
	}
}

package entity

import (
	"context"
	"time"

	"github.com/beneath-core/pkg/timeutil"
	"github.com/beneath-core/internal/hub"
	uuid "github.com/satori/go.uuid"
)

// BillingPlan represents a Billing Plan that an Organization can subscribe to
type BillingPlan struct {
	BillingPlanID          uuid.UUID       `sql:",pk,type:uuid,default:uuid_generate_v4()"`
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

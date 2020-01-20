package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
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
	SeatReadQuota          int64           `sql:",notnull"`
	SeatWriteQuota         int64           `sql:",notnull"`
	BaseReadQuota          int64           `sql:",notnull"`
	BaseWriteQuota         int64           `sql:",notnull"`
	ReadOveragePriceCents  int32           `sql:",notnull"`
	WriteOveragePriceCents int32           `sql:",notnull"`
	Personal               bool            `sql:",notnull"`
}

// FindBillingPlan finds a billing plan by ID
func FindBillingPlan(ctx context.Context, billingPlanID uuid.UUID) *BillingPlan {
	billingPlan := &BillingPlan{
		BillingPlanID: billingPlanID,
	}
	err := db.DB.ModelContext(ctx, billingPlan).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingPlan
}

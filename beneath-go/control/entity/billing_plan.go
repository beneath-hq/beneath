package entity

import (
	uuid "github.com/satori/go.uuid"
)

// BillingPlan represents a Billing Plan that an Organization can subscribe to
type BillingPlan struct {
	BillingPlanID     uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Description       string    `validate:"omitempty,lte=255"`
	PriceCurrency     string
	SeatPrice         int32   `sql:",notnull"`
	SeatReadQuota     int64   `sql:",notnull"`
	SeatWriteQuota    int64   `sql:",notnull"`
	ReadOveragePrice  float32 `sql:",notnull"`
	WriteOveragePrice float32 `sql:",notnull"`
	ExtraQuota        int64   `sql:",notnull"`
	// ModelMinuteQuota        int32   `sql:",notnull"`
	// ModelMinuteOveragePrice float32 `sql:",notnull"`
}

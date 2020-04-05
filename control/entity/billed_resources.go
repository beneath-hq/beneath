package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-org/beneath/internal/hub"
)

// BilledResource represents a resource that an organization used during the past billing period
type BilledResource struct {
	BilledResourceID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID   uuid.UUID `sql:",type:uuid,notnull"`
	BillingTime      time.Time `sql:",notnull"`
	EntityID         uuid.UUID `sql:",type:uuid,notnull"`
	EntityName       string    `sql:",notnull"`
	EntityKind       Kind      `sql:",notnull"`
	StartTime        time.Time `sql:",notnull"`
	EndTime          time.Time `sql:",notnull"`
	Product          Product   `sql:",notnull"`
	Quantity         int64     `sql:",notnull"`
	TotalPriceCents  int32     `sql:",notnull"`
	Currency         Currency  `sql:",notnull"`
	CreatedOn        time.Time `sql:",notnull,default:now()"`
	UpdatedOn        time.Time `sql:",notnull,default:now()"`
}

// FindBilledResources returns the matching billed resources or nil
func FindBilledResources(ctx context.Context, organizationID uuid.UUID, billingTime time.Time) []*BilledResource {
	var billedResources []*BilledResource
	err := hub.DB.ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time = ?", billingTime).
		Select()
	if err != nil {
		panic(err)
	}
	return billedResources
}

// CreateOrUpdateBilledResources writes the billed resources to Postgres
func CreateOrUpdateBilledResources(ctx context.Context, billedResources []*BilledResource) error {
	// specifically, do not overwrite the "created_on" field, so we can spot idempotency
	_, err := hub.DB.ModelContext(ctx, &billedResources).
		OnConflict(`(billing_time, organization_id, entity_id, product) DO UPDATE
			SET (start_time, end_time, quantity, total_price_cents, currency, updated_on) = 
			(EXCLUDED.start_time, EXCLUDED.end_time, EXCLUDED.quantity, EXCLUDED.total_price_cents, EXCLUDED.currency, EXCLUDED.updated_on)`).
		Insert()
	if err != nil {
		return err
	}

	return nil
}

// GetTotalPriceCents implements payments/driver.BilledResource
func (br *BilledResource) GetTotalPriceCents() int32 {
	return br.TotalPriceCents
}

// GetStartTime implements payments/driver.BilledResource
func (br *BilledResource) GetStartTime() time.Time {
	return br.StartTime
}

// GetEndTime implements payments/driver.BilledResource
func (br *BilledResource) GetEndTime() time.Time {
	return br.EndTime
}

// GetBillingTime implements payments/driver.BilledResource
func (br *BilledResource) GetBillingTime() time.Time {
	return br.BillingTime
}

// GetProduct implements payments/driver.BilledResource
func (br *BilledResource) GetProduct() string {
	return string(br.Product)
}

// GetEntityName implements payments/driver.BilledResource
func (br *BilledResource) GetEntityName() string {
	return string(br.EntityName)
}

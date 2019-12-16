package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/db"
	uuid "github.com/satori/go.uuid"
)

// BilledResource represents a resource that an organization used during the past billing period
type BilledResource struct {
	BilledResourceID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID   uuid.UUID `sql:",type:uuid,notnull"`
	BillingTime      time.Time `sql:",notnull"`
	EntityID         uuid.UUID `sql:",type:uuid,notnull"`
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
	err := db.DB.ModelContext(ctx, &billedResources).
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
	for _, br := range billedResources {
		// Q: this overwrites the "created_on" field, so we don't know if the billedResource was newly created or updated
		// Q: we could do this in bulk if we can do ModelContext(ctx, br1, br2, br3...)
		_, err := db.DB.ModelContext(ctx, br).
			OnConflict("(billing_time, organization_id, entity_id, product) DO UPDATE").
			Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

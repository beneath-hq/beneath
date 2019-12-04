package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/log"
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
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // defer rollback on error

	// TODO(review): Much easier and faster to do a bulk upsert, see https://github.com/go-pg/pg/issues/609

	for _, line := range billedResources {
		// query for existance
		billedResource := &BilledResource{}
		create := false
		err := tx.Model(billedResource).
			Where("organization_id = ?", line.OrganizationID).
			Where("billing_time = ?", line.BillingTime).
			Where("entity_id = ?", line.EntityID).
			Where("product = ?", line.Product).
			Where("start_time = ?", line.StartTime).
			Where("end_time = ?", line.EndTime).
			For("UPDATE").Select()

		if !AssertFoundOne(err) {
			create = true
		}

		// update
		if !create {
			billedResource.StartTime = line.StartTime
			billedResource.EndTime = line.EndTime
			billedResource.Product = line.Product
			billedResource.Quantity = line.Quantity
			billedResource.TotalPriceCents = line.TotalPriceCents
			billedResource.Currency = line.Currency
			billedResource.UpdatedOn = time.Now()
			err = tx.Update(billedResource)
			if err != nil {
				log.S.Infow("Error! ", err)
			}
		}

		// create
		if create {
			_, err := tx.ModelContext(ctx, line).Insert()
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/internal/hub"
)

// BillingMethod represents an organization's method of payment
type BillingMethod struct {
	BillingMethodID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID  uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization    *Organization
	PaymentsDriver  PaymentsDriver `sql:",notnull"`
	DriverPayload   map[string]interface{}
	CreatedOn       time.Time `sql:",default:now()"`
	UpdatedOn       time.Time `sql:",default:now()"`
}

// FindBillingMethod finds a billing method by ID
func FindBillingMethod(ctx context.Context, billingMethodID uuid.UUID) *BillingMethod {
	billingMethod := &BillingMethod{
		BillingMethodID: billingMethodID,
	}
	err := hub.DB.ModelContext(ctx, billingMethod).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingMethod
}

// FindBillingMethodsByOrganization finds all billing methods for an organization
func FindBillingMethodsByOrganization(ctx context.Context, organizationID uuid.UUID) []*BillingMethod {
	var billingMethods []*BillingMethod
	err := hub.DB.ModelContext(ctx, &billingMethods).
		Where("organization_id = ?", organizationID).
		Select()
	if err != nil {
		panic(err)
	}
	return billingMethods
}

// CreateBillingMethod creates a billing method for the given organization
func CreateBillingMethod(ctx context.Context, organizationID uuid.UUID, paymentsDriver PaymentsDriver, driverPayload map[string]interface{}) (*BillingMethod, error) {
	bm := &BillingMethod{
		OrganizationID: organizationID,
		PaymentsDriver: paymentsDriver,
		DriverPayload:  driverPayload,
	}

	err := hub.DB.Insert(bm)
	if err != nil {
		return nil, err
	}

	return bm, nil
}

package entity

import (
	"context"
	"time"

	"github.com/beneath-core/internal/hub"
	uuid "github.com/satori/go.uuid"
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

// FindBillingMethodByOrganizationAndDriver finds a billing method by organization and driver
// TODO: currently assuming that an organization does not have more than one card
func FindBillingMethodByOrganizationAndDriver(ctx context.Context, organizationID uuid.UUID, paymentDriver PaymentsDriver) *BillingMethod {
	bm := &BillingMethod{}
	err := hub.DB.Model(bm).
		Where("organization_id = ?", organizationID).
		Where("payments_driver = ?", paymentDriver).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return bm
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

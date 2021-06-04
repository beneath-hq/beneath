package billing

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/infra/db"
)

// FindBillingMethod finds a billing method
func (s *Service) FindBillingMethod(ctx context.Context, billingMethodID uuid.UUID) *models.BillingMethod {
	billingMethod := &models.BillingMethod{
		BillingMethodID: billingMethodID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, billingMethod).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return billingMethod
}

// FindBillingMethodsByOrganization finds all billing methods for an organization
func (s *Service) FindBillingMethodsByOrganization(ctx context.Context, organizationID uuid.UUID) []*models.BillingMethod {
	var billingMethods []*models.BillingMethod
	err := s.DB.GetDB(ctx).ModelContext(ctx, &billingMethods).
		Where("organization_id = ?", organizationID).
		Select()
	if err != nil {
		panic(err)
	}
	return billingMethods
}

// CreateBillingMethod creates a billing method for the given organization
func (s *Service) CreateBillingMethod(ctx context.Context, organizationID uuid.UUID, paymentsDriver string, driverPayload map[string]interface{}) (*models.BillingMethod, error) {
	bm := &models.BillingMethod{
		OrganizationID: organizationID,
		PaymentsDriver: paymentsDriver,
		DriverPayload:  driverPayload,
	}

	err := s.DB.GetDB(ctx).Insert(bm)
	if err != nil {
		return nil, err
	}

	return bm, nil
}

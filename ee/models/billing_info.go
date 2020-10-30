package models

import (
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/models"
)

// BillingInfo binds together an organization, it's chosen billing plan and it's chosen billing method
type BillingInfo struct {
	_msgpack        struct{}             `msgpack:",omitempty"`
	BillingInfoID   uuid.UUID            `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID  uuid.UUID            `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization    *models.Organization `msgpack:"-"`
	BillingMethodID *uuid.UUID           `sql:"on_delete:RESTRICT,type:uuid"`
	BillingMethod   *BillingMethod       `msgpack:"-"`
	BillingPlanID   uuid.UUID            `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingPlan     *BillingPlan         `msgpack:"-"`
	Country         string
	Region          string
	CompanyName     string
	TaxNumber       string
	CreatedOn       time.Time `sql:",default:now()"`
	UpdatedOn       time.Time `sql:",default:now()"`
}

// GetOrganizationID implements payments/driver.BillingInfo
func (bi *BillingInfo) GetOrganizationID() uuid.UUID {
	return bi.OrganizationID
}

// GetBillingPlanCurrency implements payments/driver.BillingInfo
func (bi *BillingInfo) GetBillingPlanCurrency() string {
	return string(bi.BillingPlan.Currency)
}

// GetDriverPayload implements payments/driver.BillingInfo
func (bi *BillingInfo) GetDriverPayload() map[string]interface{} {
	return bi.BillingMethod.DriverPayload
}

// GetPaymentsDriver implements payments/driver.BillingInfo
func (bi *BillingInfo) GetPaymentsDriver() string {
	return string(bi.BillingMethod.PaymentsDriver)
}

// GetCountry implements payments/driver.BillingInfo
func (bi *BillingInfo) GetCountry() string {
	return bi.Country
}

// GetRegion implements payments/driver.BillingInfo
func (bi *BillingInfo) GetRegion() string {
	return bi.Region
}

// IsCompany implements payments/driver.BillingInfo
func (bi *BillingInfo) IsCompany() bool {
	if bi.CompanyName != "" {
		return true
	}
	return false
}

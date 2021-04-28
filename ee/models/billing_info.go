package models

import (
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
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
	NextBillingTime time.Time
	LastInvoiceTime time.Time
	CreatedOn       time.Time `sql:",default:now()"`
	UpdatedOn       time.Time `sql:",default:now()"`
}

// IsCompany returns true if the organization should be treated as a company for tax purposes
func (bi *BillingInfo) IsCompany() bool {
	return bi.TaxNumber != ""
}

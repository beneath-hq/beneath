package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// EntityKind represents a resource
type EntityKind string

// Constants for EntityKind
const (
	OrganizationEntityKind EntityKind = "organization"
	ProjectEntityKind                 = "project"
	SecretEntityKind                  = "secret"
	ServiceEntityKind                 = "service"
	StreamEntityKind                  = "stream"
	UserEntityKind                    = "user"
)

// Product represents a product that we may bill for
type Product string

const (
	// SeatProduct represents the seat product
	SeatProduct Product = "seat"

	// SeatProratedProduct is like seat, but when added mid-period (eg. due to plan upgrade or new user added to an org)
	SeatProratedProduct Product = "seat_prorated"

	// BaseQuotaProduct represents the base quota product
	BaseQuotaProduct Product = "base_quota"

	// BaseQuotaProratedProduct is like BaseQuotaProduct, but when added mid-period (eg. due to a plan upgrade)
	BaseQuotaProratedProduct Product = "base_quota_prorated"

	// ReadOverageProduct represents the read_overage product
	ReadOverageProduct Product = "read_overage"

	// WriteOverageProduct represents the write_overage product
	WriteOverageProduct Product = "write_overage"

	// ScanOverageProduct represents the scan_overage product
	ScanOverageProduct Product = "scan_overage"
)

// BilledResource represents a resource that an organization used in a past billing period
type BilledResource struct {
	_msgpack         struct{}   `msgpack:",omitempty"`
	BilledResourceID uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID   uuid.UUID  `sql:",type:uuid,notnull"`
	BillingTime      time.Time  `sql:",notnull"`
	EntityID         uuid.UUID  `sql:",type:uuid,notnull"`
	EntityKind       EntityKind `sql:",notnull"`
	StartTime        time.Time  `sql:",notnull"`
	EndTime          time.Time  `sql:",notnull"`
	Product          Product    `sql:",notnull"`
	Quantity         float32    `sql:",notnull"`
	TotalPriceCents  int32      `sql:",notnull"`
	Currency         Currency   `sql:",notnull"`
	CreatedOn        time.Time  `sql:",notnull,default:now()"`
	UpdatedOn        time.Time  `sql:",notnull,default:now()"`
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

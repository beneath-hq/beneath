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
	UserEntityKind                    = "user"
)

// Product represents a product that we may bill for
type Product string

const (
	// BaseProduct represents the base price of a plan
	BaseProduct Product = "base"

	// SeatProduct represents the seat price of a plan
	SeatProduct Product = "seat"

	// ProratedSeatProduct represents a seat added mid-period
	ProratedSeatProduct Product = "prorated_seat"

	// ReadOverageProduct represents consumed read overage
	ReadOverageProduct Product = "read_overage"

	// WriteOverageProduct represents consumed write overage
	WriteOverageProduct Product = "write_overage"

	// ScanOverageProduct represents consumed scan overage
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
	Quantity         float64    `sql:",notnull"`
	TotalPriceCents  int64      `sql:",notnull"`
	Currency         Currency   `sql:",notnull"`
	CreatedOn        time.Time  `sql:",notnull,default:now()"`
	UpdatedOn        time.Time  `sql:",notnull,default:now()"`
}

// ---------------
// Events

// ShouldInvoiceBilledResourcesEvent is sent when billed resources were added AND it's time they were invoiced
type ShouldInvoiceBilledResourcesEvent struct {
	OrganizationID uuid.UUID
}

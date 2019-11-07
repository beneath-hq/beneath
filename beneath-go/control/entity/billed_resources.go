package entity

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// BilledResource represents a resource that an organization used during the past billing period
type BilledResource struct {
	OrganizationID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BillingTime      time.Time `sql:",pk"`
	BilledEntityID   uuid.UUID `sql:",pk,type:uuid"`
	BilledEntityKind BilledEntityKind
	StartTime        time.Time
	EndTime          time.Time
	ResourceID       uuid.UUID `sql:",type:uuid"`
	Quantity         int64
	TotalPrice       float32
	PriceCurrency    Currency
	InsertedOn       time.Time `sql:",default:now()"`
}

// BilledEntityKind represents a kind of billed entity -- user, service, or null
type BilledEntityKind string

const (
	// BilledEntityKindUser is a user
	BilledEntityKindUser BilledEntityKind = "u"

	// BilledEntityKindService is a service
	BilledEntityKindService BilledEntityKind = "s"

	// BilledEntityKindNull is used for invoice items that apply to the whole organization (e.g. "ExtraQuota")
	BilledEntityKindNull BilledEntityKind = ""
)

// Currency represents the currency by which the organization is billed
type Currency string

const (
	// CurrencyDollar is USD
	CurrencyDollar Currency = "usd"

	// CurrencyEuro is EUR
	CurrencyEuro Currency = "eur"
)

package models

import (
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
)

// PaymentMethod represents different methods of payment
type PaymentMethod string

// Constants for PaymentMethod
const (
	CardPaymentMethod PaymentMethod = "card"
	WirePaymentMethod               = "wire"
)

// BillingMethod represents an organization's method of payment
type BillingMethod struct {
	_msgpack        struct{}               `msgpack:",omitempty"`
	BillingMethodID uuid.UUID              `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID  uuid.UUID              `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization    *models.Organization   `msgpack:"-"`
	PaymentsDriver  string                 `sql:",notnull"`
	DriverPayload   map[string]interface{} ``
	CreatedOn       time.Time              `sql:",default:now()"`
	UpdatedOn       time.Time              `sql:",default:now()"`
}

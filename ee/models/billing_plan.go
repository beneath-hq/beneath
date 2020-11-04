package models

import (
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// Currency represents a billing currency
type Currency string

// Constants for Currency
const (
	DollarCurrency Currency = "USD"
	EuroCurrency            = "EUR"
)

// BillingPlan represents a billing plan that an organization can subscribe to.
//
// Reads:  The number of bytes that are queried from the {log, index, resulting table of a warehouse
//         query} and sent over the network.
// Writes: The number of bytes that are written to Beneath and sent over the network. We measure the
//         size of the compressed Avro data. Though we store the data in multiple downstream formats,
//         we don't multiply or split out the Write metric by the destinations.
// Scans:  The number of bytes that are scanned in a warehouse query.
//
// XQuota, where X = {Read, Write, Scan}: The absolute limits on an organization's usage. Further usage
//                                        will get shut off once quota is reached.
//
// Here's how the quota arithmetic works:
//   PrepaidXQuota = BaseXQuota + (numSeats * SeatXQuota)
//   Maximum potential XOverageBytes = XQuota - PrepaidXQuota
//
// The PrepaidXQuota is stored directly in the organization object (see organization.go)
// The XOverageBytes is computed in billing_task_b_compute_bill_resources.go
type BillingPlan struct {
	_msgpack               struct{}        `msgpack:",omitempty"`
	BillingPlanID          uuid.UUID       `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Default                bool            `sql:",notnull,default:false"`
	Description            string          `validate:"omitempty,lte=255"`
	CreatedOn              time.Time       `sql:",notnull,default:now()"`
	UpdatedOn              time.Time       `sql:",notnull,default:now()"`
	Currency               Currency        `sql:",notnull"`
	Period                 timeutil.Period `sql:",notnull"`
	BasePriceCents         int64           `sql:",notnull"`
	SeatPriceCents         int64           `sql:",notnull"`
	BaseReadQuota          int64           `sql:",notnull"` // bytes
	BaseWriteQuota         int64           `sql:",notnull"` // bytes
	BaseScanQuota          int64           `sql:",notnull"` // bytes
	SeatReadQuota          int64           `sql:",notnull"` // bytes
	SeatWriteQuota         int64           `sql:",notnull"` // bytes
	SeatScanQuota          int64           `sql:",notnull`  // bytes
	ReadQuota              int64           `sql:",notnull"` // bytes
	WriteQuota             int64           `sql:",notnull"` // bytes
	ScanQuota              int64           `sql:",notnull"` // bytes
	ReadOveragePriceCents  int64           `sql:",notnull"` // price per GB overage
	WriteOveragePriceCents int64           `sql:",notnull"` // price per GB overage
	ScanOveragePriceCents  int64           `sql:",notnull"` // price per GB overage
	MultipleUsers          bool            `sql:",notnull"`
	AvailableInUI          bool            `sql:",notnull,default:false"`
}

// IsFree returns true if the billing plan can never need payment
func (bp *BillingPlan) IsFree() bool {
	return (bp.BasePriceCents == 0 &&
		bp.SeatPriceCents == 0 &&
		bp.ReadOveragePriceCents == 0 &&
		bp.WriteOveragePriceCents == 0 &&
		bp.ScanOveragePriceCents == 0)
}

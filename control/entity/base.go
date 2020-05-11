package entity

// Kind represents a kind of entity
type Kind string

const (
	// ModelEntityKind represents a model entity
	ModelEntityKind Kind = "model"

	// OrganizationEntityKind represents a organization entity
	OrganizationEntityKind Kind = "organization"

	// ProjectEntityKind represents a project entity
	ProjectEntityKind Kind = "project"

	// SecretEntityKind represents a secret entity
	SecretEntityKind Kind = "secret"

	// ServiceEntityKind represents a service entity
	ServiceEntityKind Kind = "service"

	// StreamEntityKind represents a stream entity
	StreamEntityKind Kind = "stream"

	// UserEntityKind represents a user entity
	UserEntityKind Kind = "user"
)

// Product represents a product that we bill for
type Product string

const (
	// PrepaidQuotaProduct represents the prepaid quota product
	PrepaidQuotaProduct Product = "prepaid_quota"

	// SeatProduct represents the seat product
	SeatProduct Product = "seat"

	// PrepaidQuotaProratedProduct represents the prepaid_quota product when it is added to a bill mid-period
	// occurs when a billing plan is upgraded mid-month
	PrepaidQuotaProratedProduct Product = "prepaid_quota_prorated"

	// SeatProratedProduct represents the seat product when it is added to a bill mid-period
	// occurs when a user is added to an organization mid-month
	// also occurs when a billing plan is upgraded mid-month
	SeatProratedProduct Product = "seat_prorated"

	// ReadOverageProduct represents the read_overage product
	ReadOverageProduct Product = "read_overage"

	// WriteOverageProduct represents the write_overage product
	WriteOverageProduct Product = "write_overage"
)

// Currency represents the currency by which the organization is billed
type Currency string

const (
	// DollarCurrency is USD
	DollarCurrency Currency = "USD"

	// EuroCurrency is EUR
	EuroCurrency Currency = "EUR"
)

// PaymentMethodType represents
type PaymentMethodType string

const (
	// CardPaymentMethod means the organization's credit/debit card will be charged automatically
	CardPaymentMethod PaymentMethodType = "card"

	// WirePaymentMethod means the organization will pay via wire
	WirePaymentMethod PaymentMethodType = "wire"
)

// PaymentsDriver represents the different ways an organization can pay its bill
type PaymentsDriver string

const (
	// StripeCardDriver means the organization's credit/debit card will be charged automatically
	StripeCardDriver PaymentsDriver = "stripecard"

	// StripeWireDriver means the organization will pay via wire
	StripeWireDriver PaymentsDriver = "stripewire"

	// AnarchismDriver means the organization doesn't pay!!!
	AnarchismDriver PaymentsDriver = "anarchism"
)

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
	// SeatProduct represents the seat product
	SeatProduct Product = "seat"

	// ReadProduct represents the read product
	ReadProduct Product = "read"

	// WriteProduct represents the write product
	WriteProduct Product = "write"

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

// PaymentMethod represents
type PaymentMethod string

const (
	// CardPaymentMethod means the organization's credit/debit card will be charged automatically
	CardPaymentMethod PaymentMethod = "card"

	// WirePaymentMethod means the organization will pay via wire
	WirePaymentMethod PaymentMethod = "wire"
)

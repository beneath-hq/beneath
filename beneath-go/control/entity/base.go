package entity

// Kind represents a kind of entity
type Kind string

const (
	// ModelEntityKind represents a model entity
	ModelEntityKind = "model"

	// OrganizationEntityKind represents a organization entity
	OrganizationEntityKind = "organization"

	// ProjectEntityKind represents a project entity
	ProjectEntityKind = "project"

	// SecretEntityKind represents a secret entity
	SecretEntityKind = "secret"

	// ServiceEntityKind represents a service entity
	ServiceEntityKind = "service"

	// StreamEntityKind represents a stream entity
	StreamEntityKind = "stream"

	// UserEntityKind represents a user entity
	UserEntityKind = "user"
)

// Product represents a product that we bill for
type Product string

const (
	// SeatProduct represents the seat product
	SeatProduct = "seat"

	// ReadProduct represents the read product
	ReadProduct = "read"

	// WriteProduct represents the write product
	WriteProduct = "write"

	// ReadOverageProduct represents the read_overage product
	ReadOverageProduct = "read_overage"

	// WriteOverageProduct represents the write_overage product
	WriteOverageProduct = "write_overage"
)

// Currency represents the currency by which the organization is billed
type Currency string

const (
	// CurrencyDollar is USD
	CurrencyDollar Currency = "USD"

	// CurrencyEuro is EUR
	CurrencyEuro Currency = "EUR"
)

// PaymentMethod represents
type PaymentMethod string

const (
	// PaymentMethodCard means the organization's credit/debit card will be charged automatically
	PaymentMethodCard PaymentMethod = "card"

	// PaymentMethodWire means the organization will pay via wire
	PaymentMethodWire PaymentMethod = "wire"
)

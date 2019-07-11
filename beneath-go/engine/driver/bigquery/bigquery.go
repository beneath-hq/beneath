package bigquery

import "github.com/beneath-core/beneath-go/core"

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
}

// BigQuery implements beneath.WarehouseDriver
type BigQuery struct {
	name string
}

// New returns a new
func New() *BigQuery {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigquery", &config)

	// create instance
	b := &BigQuery{}
	b.name = "bigquery"
	return b
}

// GetMaxDataSize implements beneath.WarehouseDriver
func (p *BigQuery) GetMaxDataSize() int {
	return 1000000
}

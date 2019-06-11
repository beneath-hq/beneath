package bigtable

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID  string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID string `envconfig:"INSTANCE_ID" required:"true"`
}

// Bigtable contains the connection info for the platform
type Bigtable struct {
	name string
}

// New returns a new
func New() *Bigtable {
	// parse config from env
	var config configSpecification
	err := envconfig.Process("beneath_bigtable", &config)
	if err != nil {
		log.Fatalf("bigtable: %s", err.Error())
	}

	// create instance
	p := &Bigtable{}
	p.name = "bigtable"
	return p
}

// GetName identifies the platform
func (p *Bigtable) GetName() string {
	return p.name
}
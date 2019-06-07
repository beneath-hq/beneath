package pubsub

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

// Pubsub contains the connection info for the platform
type Pubsub struct {
	name string
}

// New returns a new
func New() *Pubsub {
	// parse config from env
	var config configSpecification
	err := envconfig.Process("beneath_bigtable", &config)
	if err != nil {
		log.Fatalf("pubsub: %s", err.Error())
	}

	// create instance
	p := &Pubsub{}
	p.name = "pubsub"
	return p
}

// GetName identifies the platform
func (p *Pubsub) GetName() string {
	return p.name
}

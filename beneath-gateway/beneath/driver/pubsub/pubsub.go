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

// Pubsub implements beneath.StreamsDriver
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

// GetName implements beneath.StreamsDriver
func (p *Pubsub) GetName() string {
	return p.name
}

// GetMaxMessageSize implements beneath.StreamsDriver
func (p *Pubsub) GetMaxMessageSize() int {
	return 10000000
}

// PushWriteRequest implements beneath.StreamsDriver
func (p *Pubsub) PushWriteRequest(data []byte) error {
	// TODO
	return nil
}

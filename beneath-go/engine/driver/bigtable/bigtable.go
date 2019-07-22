package bigtable

import "github.com/beneath-core/beneath-go/core"

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID    string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID   string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost string `envconfig:"EMULATOR_HOST" required:"false"`
}

// Bigtable implements beneath.TablesDriver
type Bigtable struct {
	name string
}

// New returns a new
func New() *Bigtable {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigtable", &config)

	// signal that you are using the BigTable emulator by setting the BIGTABLE_EMULATOR_HOST environment variable
	// once the emulator is running, run "$(gcloud beta emulators bigtable env-init)" in a terminal to set the environment variable

	// create instance
	b := &Bigtable{}
	b.name = "bigtable"
	return b
}

// GetName implements beneath.TablesDriver
func (p *Bigtable) GetName() string {
	return p.name
}

// GetMaxKeySize implements beneath.TablesDriver
func (p *Bigtable) GetMaxKeySize() int {
	return 2048
}

// GetMaxDataSize implements beneath.TablesDriver
func (p *Bigtable) GetMaxDataSize() int {
	return 1000000
}

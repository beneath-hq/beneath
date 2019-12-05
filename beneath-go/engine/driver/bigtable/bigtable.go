package bigtable

import (
	"context"
	"os"
	"sync"

	"github.com/beneath-core/beneath-go/engine/driver"

	"cloud.google.com/go/bigtable"
	"github.com/beneath-core/beneath-go/core"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
	ProjectID    string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID   string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost string `envconfig:"EMULATOR_HOST" required:"false"`
}

// BigTable implements beneath.Log and beneath.LookupService
type BigTable struct {
	Admin   *bigtable.AdminClient
	Client  *bigtable.Client
	Records *bigtable.Table
	Latest  *bigtable.Table
	Metrics *bigtable.Table
}

// Global
var global BigTable
var once sync.Once

func createGlobal() {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_bigtable", &config)

	// if EMULATOR_HOST set, configure bigtable for the emulator
	if config.EmulatorHost != "" {
		os.Setenv("BIGTABLE_PROJECT_ID", config.ProjectID)
		os.Setenv("BIGTABLE_EMULATOR_HOST", config.EmulatorHost)
	}

	// create BigTable admin
	admin, err := bigtable.NewAdminClient(context.Background(), config.ProjectID, config.InstanceID)
	if err != nil {
		panic(err)
	}

	// prepare BigTable client
	client, err := bigtable.NewClient(context.Background(), config.ProjectID, config.InstanceID)
	if err != nil {
		panic(err)
	}

	// create instance
	global = BigTable{
		Admin:  admin,
		Client: client,
	}
}

// GetLog returns a BigTable implementation of beneath.Log
func GetLog() driver.Log {
	once.Do(createGlobal)
	return global
}

// GetLookupService returns a BigTable implementation of beneath.LookupService
func GetLookupService() driver.LookupService {
	once.Do(createGlobal)
	return global
}

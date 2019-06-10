package beneath

import (
	"log"

	"github.com/beneath-core/beneath-gateway/pkg/beneath/platform/bigtable"
	"github.com/beneath-core/beneath-gateway/pkg/beneath/platform/pubsub"
)

// StreamsPlatform defines the functions necessary to encapsulate Beneath's streaming data needs
type StreamsPlatform interface {
	GetName() string
	// TODO: Add streaming data functions
}

// TablesPlatform defines the functions necessary to encapsulate Beneath's operational datastore needs
type TablesPlatform interface {
	GetName() string
	// TODO: Add table functions
}

var (
	// Streams is the global instance of StreamsPlatform
	Streams StreamsPlatform

	// Tables is the global instance of TablesPlatform
	Tables TablesPlatform
)

func init() {
	// init Streams
	switch Config.StreamsPlatform {
	case "pubsub":
		Streams = pubsub.New()
	default:
		log.Fatalf("invalid streams platform %s", Config.StreamsPlatform)
	}

	// init Tables
	switch Config.TablesPlatform {
	case "bigtable":
		Tables = bigtable.New()
	default:
		log.Fatalf("invalid tables platform %s", Config.TablesPlatform)
	}

	// // Init Warehouse
	// switch Config.WarehousePlatform {
	// case "bigquery":
	// 	Warehouse = bigquery.New()
	// default:
	// 	log.Fatalf("invalid tables platform %s", Config.WarehousePlatform)
	// }
}

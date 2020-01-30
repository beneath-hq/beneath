package bigtable

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/bigtable"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/engine/driver"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable/sequencer"
)

// configSpecification defines the config variables to load from ENV
type configSpecification struct {
	ProjectID    string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID   string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost string `envconfig:"EMULATOR_HOST" required:"false"`
}

// BigTable implements beneath.Log and beneath.LookupService
type BigTable struct {
	Admin           *bigtable.AdminClient
	Client          *bigtable.Client
	Sequencer       sequencer.Sequencer
	Log             *bigtable.Table
	LogExpiring     *bigtable.Table
	Indexes         *bigtable.Table
	IndexesExpiring *bigtable.Table
	Metrics         *bigtable.Table
}

const (
	/*
		SCHEMA NOTES:
		- Tables: log and log_exp
		  - Keys: [16-byte instance ID][Tuple{int64(offset), bytes(primary_key)}]
		  - Columns:
		    - f:a -> avro data
		    - f:t -> pipeline process time in ms since epoch (big-endian int64)
		- Tables: indexes and indexes_exp
		  - Keys:
		    - For primary indexes: [16-byte index ID][primary_key]
		    - For secondary indexes: [16-byte index ID][Tuple{bytes(secondary_key), bytes(primary_key))}]
		  - Columns:
		    - If denormalized: f:a -> avro data
		    - If normalized:   f:o -> log offset (big-endian int64)
	*/

	logTableName             = "log"
	logExpiringTableName     = "log_exp"
	logColumnFamilyName      = "f"
	logAvroColumnName        = "a"
	logProcessTimeColumnName = "t"

	indexesTableName         = "indexes"
	indexesExpiringTableName = "indexes_exp"
	indexesColumnFamilyName  = "f"
	indexesAvroColumnName    = "a"
	indexesOffsetColumnName  = "o"

	sequencerTableName = "sequencer"

	metricsTableName            = "metrics"
	metricsColumnFamilyName     = "cf0"
	metricsReadOpsColumnName    = "ro"
	metricsReadRowsColumnName   = "rr"
	metricsReadBytesColumnName  = "rb"
	metricsWriteOpsColumnName   = "wo"
	metricsWriteRowsColumnName  = "wr"
	metricsWriteBytesColumnName = "wb"
)

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

	global.Sequencer, err = sequencer.New(admin, client, sequencerTableName)
	if err != nil {
		panic(err)
	}

	global.Log = global.openTable(logTableName, logColumnFamilyName, 1, false)
	global.LogExpiring = global.openTable(logExpiringTableName, logColumnFamilyName, 1, true)
	global.Indexes = global.openTable(indexesTableName, indexesColumnFamilyName, 1, false)
	global.IndexesExpiring = global.openTable(indexesExpiringTableName, indexesColumnFamilyName, 1, true)
	global.Metrics = global.openTable(metricsTableName, metricsColumnFamilyName, 1, false)
}

// GetLookupService returns a BigTable implementation of beneath.LookupService
func GetLookupService() driver.LookupService {
	once.Do(createGlobal)
	return global
}

func (b BigTable) openTable(name string, cfName string, maxVersions int, expiring bool) *bigtable.Table {
	// create table
	err := b.Admin.CreateTable(context.Background(), name)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating table '%s': %v", name, err))
		}
	}

	// create column family
	err = b.Admin.CreateColumnFamily(context.Background(), name, cfName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating column family '%s': %v", cfName, err))
		}
	}

	// set garbage collection policy
	policy := bigtable.MaxVersionsPolicy(maxVersions)
	if expiring {
		// if expiring, set timestamps to be expiration times
		policy = bigtable.UnionPolicy(policy, bigtable.MaxAgePolicy(time.Second))
	}

	err = b.Admin.SetGCPolicy(context.Background(), name, cfName, policy)
	if err != nil {
		panic(fmt.Errorf("error setting gc policy: %v", err))
	}

	// open connection
	return b.Client.Open(name)
}

func (b BigTable) tablesForStream(s driver.Stream) (*bigtable.Table, *bigtable.Table) {
	if streamExpires(s) {
		return b.LogExpiring, b.IndexesExpiring
	}
	return b.Log, b.Indexes
}

func (b BigTable) logTable(expires bool) *bigtable.Table {
	if expires {
		return b.LogExpiring
	}
	return b.Log
}

func (b BigTable) indexesTable(expires bool) *bigtable.Table {
	if expires {
		return b.IndexesExpiring
	}
	return b.Indexes
}

func (b BigTable) readLog(ctx context.Context, expires bool, rs bigtable.RowSet, limit int, cb func(bigtable.Row) bool) error {
	return b.logTable(expires).ReadRows(ctx, rs, cb, b.readFilter(expires), bigtable.LimitRows(int64(limit)))
}

func (b BigTable) readIndexes(ctx context.Context, expires bool, rs bigtable.RowSet, limit int, cb func(bigtable.Row) bool) error {
	return b.indexesTable(expires).ReadRows(ctx, rs, cb, b.readFilter(expires), bigtable.LimitRows(int64(limit)))
}

func (b BigTable) readFilter(expires bool) bigtable.ReadOption {
	// bigtable requires us to set filters that match our garbage collection policy
	// to get semantically correct results
	filter := bigtable.LatestNFilter(1)
	if expires {
		// In expiring streams, the timestamp is interpreted as the expiration date.
		// So the timestamp must be greater than now to not be expired.
		notExpiredFilter := bigtable.TimestampRangeFilter(time.Now(), time.Time{})
		filter = bigtable.ChainFilters(filter, notExpiredFilter)
	}
	return bigtable.RowFilter(filter)
}

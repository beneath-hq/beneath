package bigtable

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver/bigtable/sequencer"
)

// BigTable implements beneath.Log and beneath.LookupService
type BigTable struct {
	Admin           *bigtable.AdminClient
	Client          *bigtable.Client
	Sequencer       sequencer.Sequencer
	Log             *bigtable.Table
	LogExpiring     *bigtable.Table
	Indexes         *bigtable.Table
	IndexesExpiring *bigtable.Table
	Usage           *bigtable.Table
	UsageTemp       *bigtable.Table
}

// Options for BigTable
type Options struct {
	ProjectID    string `mapstructure:"project_id"`
	InstanceID   string `mapstructure:"instance_id"`
	EmulatorHost string `mapstructure:"emulator_host"`
}

func init() {
	driver.AddDriver("bigtable", newBigTable)
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

	usageTableName            = "usage"
	usageTempTableName        = "usage_tmp"
	usageTempRetention        = 8 * 24 * time.Hour
	usageColumnFamilyName     = "f"
	usageReadOpsColumnName    = "ro"
	usageReadRowsColumnName   = "rr"
	usageReadBytesColumnName  = "rb"
	usageWriteOpsColumnName   = "wo"
	usageWriteRowsColumnName  = "wr"
	usageWriteBytesColumnName = "wb"
	usageScanOpsColumnName    = "so"
	usageScanBytesColumnName  = "sb"
)

func newBigTable(optsMap map[string]interface{}) (driver.Service, error) {
	// load options
	var opts Options
	err := mapstructure.Decode(optsMap, &opts)
	if err != nil {
		return nil, fmt.Errorf("error decoding bigtable options: %s", err.Error())
	}

	// if EMULATOR_HOST set, configure bigtable for the emulator
	if opts.EmulatorHost != "" {
		os.Setenv("BIGTABLE_PROJECT_ID", opts.ProjectID)
		os.Setenv("BIGTABLE_EMULATOR_HOST", opts.EmulatorHost)
	}

	// create BigTable admin
	admin, err := bigtable.NewAdminClient(context.Background(), opts.ProjectID, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	// prepare BigTable client
	client, err := bigtable.NewClient(context.Background(), opts.ProjectID, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	// create instance
	bt := &BigTable{
		Admin:  admin,
		Client: client,
	}

	bt.Sequencer, err = sequencer.New(admin, client, sequencerTableName)
	if err != nil {
		return nil, err
	}

	// expiring tables have max age of 1 second, making their timestamps expiration dates

	bt.Log = bt.openTable(logTableName, logColumnFamilyName, 1, 0)
	bt.LogExpiring = bt.openTable(logExpiringTableName, logColumnFamilyName, 1, time.Second)
	bt.Indexes = bt.openTable(indexesTableName, indexesColumnFamilyName, 1, 0)
	bt.IndexesExpiring = bt.openTable(indexesExpiringTableName, indexesColumnFamilyName, 1, time.Second)
	bt.Usage = bt.openTable(usageTableName, usageColumnFamilyName, 1, 0)
	bt.UsageTemp = bt.openTable(usageTempTableName, usageColumnFamilyName, 1, usageTempRetention)

	return bt, nil
}

// AsLookupService implements Service
func (b *BigTable) AsLookupService() driver.LookupService {
	return b
}

// AsWarehouseService implements Service
func (b *BigTable) AsWarehouseService() driver.WarehouseService {
	return nil
}

// AsUsageService implements Service
func (b *BigTable) AsUsageService() driver.UsageService {
	return b
}

func (b BigTable) openTable(name string, cfName string, maxVersions int, maxAge time.Duration) *bigtable.Table {
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
	if maxAge != 0 {
		policy = bigtable.UnionPolicy(policy, bigtable.MaxAgePolicy(maxAge))
	}

	err = b.Admin.SetGCPolicy(context.Background(), name, cfName, policy)
	if err != nil {
		panic(fmt.Errorf("error setting gc policy: %v", err))
	}

	// open connection
	return b.Client.Open(name)
}

func (b BigTable) tablesForStream(s driver.Stream) (logTable *bigtable.Table, indexTable *bigtable.Table) {
	if logExpires(s) {
		logTable = b.LogExpiring
	} else {
		logTable = b.Log
	}

	if indexExpires(s) {
		indexTable = b.IndexesExpiring
	} else {
		indexTable = b.Indexes
	}

	return logTable, indexTable
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

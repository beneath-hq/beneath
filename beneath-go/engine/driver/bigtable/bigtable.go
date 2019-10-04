package bigtable

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigtable"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	pb "github.com/beneath-core/beneath-go/proto"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID    string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID   string `envconfig:"INSTANCE_ID" required:"true"`
	EmulatorHost string `envconfig:"EMULATOR_HOST" required:"false"`
}

// Bigtable implements beneath.TablesDriver
type Bigtable struct {
	Admin   *bigtable.AdminClient
	Client  *bigtable.Client
	Records *bigtable.Table
	Latest  *bigtable.Table
	Metrics *bigtable.Table
}

const (
	recordsTableName        = "records"
	recordsColumnFamilyName = "cf0"
	recordsColumnName       = "avro0"

	latestTableName        = "latest"
	latestColumnFamilyName = "cf0"
	latestColumnName       = "avro0"

	metricsTableName            = "metrics"
	metricsColumnFamilyName     = "cf0"
	metricsReadOpsColumnName    = "ro"
	metricsReadRowsColumnName   = "rr"
	metricsReadBytesColumnName  = "rb"
	metricsWriteOpsColumnName   = "wo"
	metricsWriteRowsColumnName  = "wr"
	metricsWriteBytesColumnName = "wb"

	maxLatestRecords = 1000
)

// New returns a new
func New() *Bigtable {
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
	bt := &Bigtable{
		Admin:  admin,
		Client: client,
	}

	// open tables
	bt.Records = bt.openTable(recordsTableName, recordsColumnFamilyName, 1)
	bt.Latest = bt.openTable(latestTableName, latestColumnFamilyName, maxLatestRecords)
	bt.Metrics = bt.openTable(metricsTableName, metricsColumnFamilyName, 1)

	return bt
}

// GetMaxKeySize implements engine.TablesDriver
func (p *Bigtable) GetMaxKeySize() int {
	return 2048
}

// GetMaxDataSize implements engine.TablesDriver
func (p *Bigtable) GetMaxDataSize() int {
	return 1000000
}

// WriteRecords implements engine.TablesDriver
func (p *Bigtable) WriteRecords(ctx context.Context, instanceID uuid.UUID, keys [][]byte, avroData [][]byte, timestamps []time.Time, saveLatest bool) error {
	// ensure all WriteRequest objects the same length
	if !(len(keys) == len(avroData) && len(keys) == len(timestamps)) {
		return fmt.Errorf("error: keys, data, and timestamps do not all have the same length")
	}

	muts := make([]*bigtable.Mutation, len(keys))
	rowKeys := make([]string, len(keys))

	var latestMut *bigtable.Mutation
	if saveLatest {
		latestMut = bigtable.NewMutation()
	}

	// create one mutation for each record in the WriteRequest
	for i, key := range keys {
		t := bigtable.Time(timestamps[i])

		muts[i] = bigtable.NewMutation()
		muts[i].Set(recordsColumnFamilyName, recordsColumnName, t, avroData[i])
		rowKeys[i] = string(makeRowKey(instanceID, key))

		if saveLatest {
			latestMut.Set(latestColumnFamilyName, latestColumnName, t, avroData[i])
		}
	}

	// apply all the mutations (i.e. write all the records) at once
	_, err := p.Records.ApplyBulk(ctx, rowKeys, muts)
	if err != nil {
		return err
	}

	// save latest
	if saveLatest {
		err := p.Latest.Apply(ctx, string(instanceID.Bytes()), latestMut)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadRecords implements engine.TablesDriver
func (p *Bigtable) ReadRecords(ctx context.Context, instanceID uuid.UUID, keys [][]byte, fn func(idx uint, avroData []byte, timestamp time.Time) error) error {
	// convert keys to RowList
	rl := make(bigtable.RowList, len(keys))
	for idx, key := range keys {
		rl[idx] = string(makeRowKey(instanceID, key))
	}

	// define callback triggered on each bigtable row
	var idx uint
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// column families are returned as a map, columns as an array; we're expecting one column atm, so just take index 0
		item := row[recordsColumnFamilyName][0]

		// trigger callback
		cbErr = fn(idx, item.Value, item.Timestamp.Time())
		if cbErr != nil {
			return false // stop
		}

		// continue
		idx++
		return true
	}

	// read rows
	err := p.Records.ReadRows(ctx, rl, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

// ReadRecordRange implements engine.TablesDriver
func (p *Bigtable) ReadRecordRange(ctx context.Context, instanceID uuid.UUID, keyRange codec.KeyRange, limit int, fn func(avroData []byte, timestamp time.Time) error) error {
	// convert keyRange to RowSet
	var rr bigtable.RowSet
	if keyRange.IsNil() {
		rr = bigtable.PrefixRange(string(instanceID[:]))
	} else if keyRange.CheckUnique() {
		rk := makeRowKey(instanceID, keyRange.Base)
		rr = bigtable.SingleRow(string(rk))
	} else if keyRange.RangeEnd == nil {
		bk := makeRowKey(instanceID, keyRange.Base)
		ek := tuple.PrefixSuccessor(instanceID[:])
		rr = bigtable.NewRange(string(bk), string(ek))
	} else {
		bk := makeRowKey(instanceID, keyRange.Base)
		ek := makeRowKey(instanceID, keyRange.RangeEnd)
		rr = bigtable.NewRange(string(bk), string(ek))
	}

	// define callback triggered on each bigtable row
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// column families are returned as a map, columns as an array; we're expecting one column atm, so just take index 0
		item := row[recordsColumnFamilyName][0]

		// trigger callback
		cbErr = fn(item.Value, item.Timestamp.Time())
		if cbErr != nil {
			return false // stop
		}

		// continue
		return true
	}

	// read rows
	err := p.Records.ReadRows(ctx, rr, cb, bigtable.LimitRows(int64(limit)), bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

// ReadLatestRecords implements engine.TablesDriver
func (p *Bigtable) ReadLatestRecords(ctx context.Context, instanceID uuid.UUID, limit int, before time.Time, fn func(avroData []byte, timestamp time.Time) error) error {
	// create filter
	var filter bigtable.Filter
	if before.IsZero() {
		filter = bigtable.LatestNFilter(limit)
	} else {
		filter = bigtable.ChainFilters(bigtable.TimestampRangeFilter(time.Time{}, before), bigtable.LatestNFilter(limit))
	}

	// read row
	row, err := p.Latest.ReadRow(ctx, string(instanceID.Bytes()), bigtable.RowFilter(filter))
	if err != nil {
		return err
	}

	// iterate through versions and return
	for _, item := range row[latestColumnFamilyName] {
		err := fn(item.Value, item.Timestamp.Time())
		if err != nil {
			return err
		}
	}

	// done
	return nil
}

// ClearRecords implements engine.TablesDriver
func (p *Bigtable) ClearRecords(ctx context.Context, instanceID uuid.UUID) error {
	return p.Admin.DropRowRange(ctx, recordsTableName, string(instanceID[:]))
}

// CommitUsage implements engine.TablesDriver
func (p *Bigtable) CommitUsage(ctx context.Context, key []byte, usage pb.QuotaUsage) error {
	rmw := bigtable.NewReadModifyWrite()

	if usage.ReadOps > 0 {
		rmw.Increment(metricsColumnFamilyName, metricsReadOpsColumnName, usage.ReadOps)
		rmw.Increment(metricsColumnFamilyName, metricsReadRowsColumnName, usage.ReadRecords)
		rmw.Increment(metricsColumnFamilyName, metricsReadBytesColumnName, usage.ReadBytes)
	}

	if usage.WriteOps > 0 {
		rmw.Increment(metricsColumnFamilyName, metricsWriteOpsColumnName, usage.WriteOps)
		rmw.Increment(metricsColumnFamilyName, metricsWriteRowsColumnName, usage.WriteRecords)
		rmw.Increment(metricsColumnFamilyName, metricsWriteBytesColumnName, usage.WriteBytes)
	}

	// apply
	_, err := p.Metrics.ApplyReadModifyWrite(ctx, string(key), rmw)

	return err
}

// ReadUsage implements engine.TablesDriver
func (p *Bigtable) ReadUsage(ctx context.Context, keyPrefix []byte, fn func(key []byte, usage pb.QuotaUsage) error) error {
	// define row range
	rr := bigtable.PrefixRange(string(keyPrefix))

	// add filter for latest cell value
	filter := bigtable.RowFilter(bigtable.LatestNFilter(1))

	// define callback triggered on each bigtable row
	var idx uint
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// save key
		key := []byte(row.Key())

		// get metrics
		u := pb.QuotaUsage{}

		for _, cell := range row[metricsColumnFamilyName] {
			colName := cell.Column[len(metricsColumnFamilyName)+1:]

			switch colName {
			case metricsReadOpsColumnName:
				u.ReadOps = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsReadRowsColumnName:
				u.ReadRecords = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsReadBytesColumnName:
				u.ReadBytes = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteOpsColumnName:
				u.WriteOps = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteRowsColumnName:
				u.WriteRecords = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteBytesColumnName:
				u.WriteBytes = int64(binary.BigEndian.Uint64(cell.Value))
			}
		}

		// trigger callback
		cbErr = fn(key, u)
		if cbErr != nil {
			return false // stop
		}

		// continue
		idx++
		return true
	}

	// read rows
	err := p.Metrics.ReadRows(ctx, rr, cb, filter)
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

// ReadUsageRange implements engine.TablesDriver
func (p *Bigtable) ReadUsageRange(ctx context.Context, fromKey []byte, toKey []byte, fn func(key []byte, usage pb.QuotaUsage) error) error {
	// convert keyRange to RowSet
	rr := bigtable.NewRange(string(fromKey), string(toKey))

	// define callback triggered on each bigtable row
	var idx uint
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// save key
		key := []byte(row.Key())

		// get metrics
		u := pb.QuotaUsage{}

		for _, cell := range row[metricsColumnFamilyName] {
			colName := cell.Column[len(metricsColumnFamilyName)+1:]

			switch colName {
			case metricsReadOpsColumnName:
				u.ReadOps = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsReadRowsColumnName:
				u.ReadRecords = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsReadBytesColumnName:
				u.ReadBytes = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteOpsColumnName:
				u.WriteOps = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteRowsColumnName:
				u.WriteRecords = int64(binary.BigEndian.Uint64(cell.Value))
			case metricsWriteBytesColumnName:
				u.WriteBytes = int64(binary.BigEndian.Uint64(cell.Value))
			}
		}

		// trigger callback
		cbErr = fn(key, u)
		if cbErr != nil {
			return false // stop
		}

		// continue
		idx++
		return true
	}

	// read rows
	err := p.Metrics.ReadRows(ctx, rr, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

func (p *Bigtable) openTable(name string, cfName string, maxVersions int) *bigtable.Table {
	// create table
	err := p.Admin.CreateTable(context.Background(), name)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating table '%s': %v", name, err))
		}
	}

	// create column family
	err = p.Admin.CreateColumnFamily(context.Background(), name, cfName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating column family '%s': %v", cfName, err))
		}
	}

	// set garbage collection policy
	err = p.Admin.SetGCPolicy(context.Background(), name, cfName, bigtable.MaxVersionsPolicy(maxVersions))
	if err != nil {
		panic(fmt.Errorf("error setting gc policy: %v", err))
	}

	// open connection
	return p.Client.Open(name)
}

// combine instanceID and key to a row key
func makeRowKey(instanceID uuid.UUID, key []byte) []byte {
	return append(instanceID[:], key...)
}

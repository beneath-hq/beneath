package bigtable

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/bigtable"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/codec"
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
	Client  *bigtable.Client
	Records *bigtable.Table
	Latest  *bigtable.Table
}

const (
	recordsTableName        = "records"
	recordsColumnFamilyName = "cf0"
	recordsColumnName       = "avro0"

	latestTableName        = "latest"
	latestColumnFamilyName = "cf0"
	latestColumnName       = "avro0"

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
		log.Fatalf("Could not create admin client: %v", err)
	}
	defer admin.Close()

	// initialize tables if they don't exist
	initializeRecordsTable(admin)
	initializeLatestTable(admin)

	// prepare BigTable client
	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, config.ProjectID, config.InstanceID)
	if err != nil {
		log.Fatalf("Could not create bigtable client: %v", err)
	}

	// create instance
	return &Bigtable{
		Client:  client,
		Records: client.Open(recordsTableName),
		Latest:  client.Open(latestTableName),
	}
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
func (p *Bigtable) WriteRecords(instanceID uuid.UUID, keys [][]byte, avroData [][]byte, timestamps []time.Time, saveLatest bool) error {
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
	_, err := p.Records.ApplyBulk(context.Background(), rowKeys, muts)
	if err != nil {
		return err
	}

	// save latest
	if saveLatest {
		err := p.Latest.Apply(context.Background(), string(instanceID.Bytes()), latestMut)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadRecords implements engine.TablesDriver
func (p *Bigtable) ReadRecords(instanceID uuid.UUID, keys [][]byte, fn func(idx uint, avroData []byte, timestamp time.Time) error) error {
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
	err := p.Records.ReadRows(context.Background(), rl, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

// ReadRecordRange implements engine.TablesDriver
func (p *Bigtable) ReadRecordRange(instanceID uuid.UUID, keyRange codec.KeyRange, limit int, fn func(avroData []byte, timestamp time.Time) error) error {
	// convert keyRange to RowSet
	var rr bigtable.RowSet
	if keyRange.IsNil() {
		rr = bigtable.PrefixRange(string(instanceID[:]))
	} else if keyRange.CheckUnique() {
		rk := makeRowKey(instanceID, keyRange.Base)
		rr = bigtable.SingleRow(string(rk))
	} else if keyRange.RangeEnd == nil {
		rk := makeRowKey(instanceID, keyRange.Base)
		rr = bigtable.InfiniteRange(string(rk))
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
	err := p.Records.ReadRows(context.Background(), rr, cb, bigtable.LimitRows(int64(limit)), bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

// ReadLatestRecords implements engine.TablesDriver
func (p *Bigtable) ReadLatestRecords(instanceID uuid.UUID, limit int, before time.Time, fn func(avroData []byte, timestamp time.Time) error) error {
	// create filter
	var filter bigtable.Filter
	if before.IsZero() {
		filter = bigtable.LatestNFilter(limit)
	} else {
		filter = bigtable.ChainFilters(bigtable.TimestampRangeFilter(time.Time{}, before), bigtable.LatestNFilter(limit))
	}

	// read row
	row, err := p.Latest.ReadRow(context.Background(), string(instanceID.Bytes()), bigtable.RowFilter(filter))
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

// combine instanceID and key to a row key
func makeRowKey(instanceID uuid.UUID, key []byte) []byte {
	return append(instanceID[:], key...)
}

func initializeRecordsTable(admin *bigtable.AdminClient) {
	// create table
	err := admin.CreateTable(context.Background(), recordsTableName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating table '%s': %v", recordsTableName, err)
		}
	}

	// create column family
	err = admin.CreateColumnFamily(context.Background(), recordsTableName, recordsColumnFamilyName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating column family '%s': %v", recordsColumnFamilyName, err)
		}
	}

	// set garbage collection policy
	err = admin.SetGCPolicy(context.Background(), recordsTableName, recordsColumnFamilyName, bigtable.MaxVersionsPolicy(1))
	if err != nil {
		log.Panicf("error setting gc policy: %v", err)
	}
}

func initializeLatestTable(admin *bigtable.AdminClient) {
	// create table
	err := admin.CreateTable(context.Background(), latestTableName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating table '%s': %v", latestTableName, err)
		}
	}

	// create column family
	err = admin.CreateColumnFamily(context.Background(), latestTableName, latestColumnFamilyName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating column family '%s': %v", latestColumnFamilyName, err)
		}
	}

	// set garbage collection policy
	err = admin.SetGCPolicy(context.Background(), latestTableName, latestColumnFamilyName, bigtable.MaxVersionsPolicy(maxLatestRecords))
	if err != nil {
		log.Panicf("error setting gc policy: %v", err)
	}
}

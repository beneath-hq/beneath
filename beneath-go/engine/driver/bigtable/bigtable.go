package bigtable

import (
	"context"
	"log"
	"os"

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
}

const (
	recordsTableName        = "records"
	recordsColumnFamilyName = "cf0"
	recordsColumnName       = "avro0"
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

// WriteRecord implements engine.TablesDriver
func (p *Bigtable) WriteRecord(instanceID uuid.UUID, key []byte, avroData []byte, sequenceNumber int64) error {
	mut := bigtable.NewMutation()
	mut.Set(recordsColumnFamilyName, recordsColumnName, bigtable.Timestamp(sequenceNumber*1000), avroData)

	rowKey := makeRowKey(instanceID, key)
	err := p.Records.Apply(context.Background(), string(rowKey), mut)
	if err != nil {
		return err
	}

	return nil
}

// ReadRecords implements engine.TablesDriver
func (p *Bigtable) ReadRecords(instanceID uuid.UUID, keyRange *codec.KeyRange, limit int, fn func(avroData []byte, sequenceNumber int64) error) error {
	// convert keyRange to RowSet
	var rr bigtable.RowSet
	if keyRange.Unique() {
		rk := makeRowKey(instanceID, keyRange.Base)
		rr = bigtable.SingleRow(string(rk))
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
		cbErr = fn(item.Value, int64(item.Timestamp))
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
}

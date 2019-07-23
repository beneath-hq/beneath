package bigtable

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/bigtable"
	"github.com/beneath-core/beneath-go/core"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID        string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID       string `envconfig:"INSTANCE_ID" required:"true"`
	RecordsTableName string `envconfig:"RECORDS_TABLE_NAME" required:"true"`
	EmulatorHost     string `envconfig:"EMULATOR_HOST" required:"false"`
}

// Bigtable implements beneath.TablesDriver
type Bigtable struct {
	Client *bigtable.Client
	Table  *bigtable.Table
}

const (
	columnFamilyName = "cf0"

	columnName = "avro0"
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

	// prepare BigTable client
	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, config.ProjectID, config.InstanceID)
	if err != nil {
		log.Fatalf("Could not create bigtable client: %v", err)
	}

	// create table in BigTable
	createBigTable(ctx, config.ProjectID, config.InstanceID, config.RecordsTableName)

	// open main table
	table := client.Open(config.RecordsTableName)

	// create instance
	return &Bigtable{
		Client: client,
		Table:  table,
	}
}

// GetMaxKeySize implements beneath.TablesDriver
func (p *Bigtable) GetMaxKeySize() int {
	return 2048
}

// GetMaxDataSize implements beneath.TablesDriver
func (p *Bigtable) GetMaxDataSize() int {
	return 1000000
}

// WriteRecordToTable ...
func (p *Bigtable) WriteRecordToTable(instanceID uuid.UUID, encodedKey []byte, avroData []byte, sequenceNumber int64) error {
	ctx := context.Background()

	rowKey := makeBigTableKey(instanceID, encodedKey)

	log.Printf("writing to table.")
	muts := bigtable.NewMutation()
	muts.Set(columnFamilyName, columnName, bigtable.Timestamp(sequenceNumber*1000), avroData)

	err := p.Table.Apply(ctx, string(rowKey), muts)
	if err != nil {
		return err
	}

	return nil
}

// ReadRecordsFromTable ...
func (p *Bigtable) ReadRecordsFromTable(instanceID uuid.UUID, encodedKey []byte) error {
	ctx := context.Background()

	log.Printf("Reading all rows:")
	rr := bigtable.PrefixRange("")
	err := p.Table.ReadRows(ctx, rr, func(row bigtable.Row) bool {
		item := row[columnFamilyName][0]
		log.Printf("\tKey: %s; Value: %s\n", item.Row, string(item.Value))
		log.Printf("\tKey: %s; Timestamp: %v\n", item.Row, item.Timestamp)
		return true
	})

	if err != nil {
		return err
	}

	// if an instanceID + encodedKey is provided, read the specify row
	log.Printf(instanceID.String())
	log.Printf(string(encodedKey))
	// rowKey := makeBigTableKey(instanceID, encodedKey)
	// r, err = p.Table.ReadRow(ctx, rowKey)
	// log.Print(r)

	return nil
}

// make bigtable key
func makeBigTableKey(instanceID uuid.UUID, encodedKey []byte) []byte {
	return append(instanceID[:], encodedKey...)
}

// create table in bigtable via the admin client
func createBigTable(ctx context.Context, project string, instanceid string, tablename string) error {
	adminClient, err := bigtable.NewAdminClient(ctx, project, instanceid)
	if err != nil {
		log.Fatalf("Could not create admin client: %v", err)
	}

	err = adminClient.CreateTable(ctx, tablename)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating table '%s': %v", tablename, err)
		} else {
			log.Printf("trying to create table that already exists for project '%s'", project)
		}
	}

	err = adminClient.CreateColumnFamily(ctx, tablename, columnFamilyName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			log.Panicf("error creating column family '%s': %v", columnFamilyName, err)
		} else {
			log.Printf("trying to create column family that already exists")
		}
	}

	if err = adminClient.Close(); err != nil {
		log.Fatalf("Could not close admin client: %v", err)
	}

	return nil
}

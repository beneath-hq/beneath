package bigquery

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"

	bq "cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
)

// InternalRow represents a record to be saved for internal future use (i.e., not for customers to see)
type InternalRow struct {
	Key       []byte
	Data      []byte
	Timestamp time.Time
}

var (
	internalRowSchema bq.Schema
)

func init() {
	internalRowSchema = bq.Schema{
		{Name: "key", Required: true, Type: bq.BytesFieldType},
		{Name: "data", Repeated: false, Type: bq.BytesFieldType},
		{Name: "timestamp", Required: true, Type: bq.TimestampFieldType},
	}
}

// ExternalRow represents a record saved for external use (i.e. with columns matching schema)
// It implements bigquery.ValueSaver
type ExternalRow struct {
	Data     map[string]interface{}
	InsertID string
}

// Save implements bigquery.ValueSaver
func (r *ExternalRow) Save() (row map[string]bq.Value, insertID string, err error) {
	data := make(map[string]bq.Value, len(r.Data))
	for k, v := range r.Data {
		data[k] = r.recursiveSerialize(v)
	}
	return data, r.InsertID, nil
}

// the bigquery client serializes every type correctly except big numbers and byte arrays;
// we handle those by a recursive search
// note: overrides in place
func (r *ExternalRow) recursiveSerialize(valT interface{}) bq.Value {
	switch val := valT.(type) {
	case *big.Int:
		return val.String()
	case *big.Rat:
		return val.FloatString(0)
	case []byte:
		return "0x" + hex.EncodeToString(val)
	case map[string]interface{}:
		for k, v := range val {
			val[k] = r.recursiveSerialize(v)
		}
	case []interface{}:
		for i, v := range val {
			val[i] = r.recursiveSerialize(v)
		}
	}
	return valT
}

// GetMaxDataSize implements engine.WarehouseDriver
func (b *BigQuery) GetMaxDataSize() int {
	return 1000000
}

// WriteRecords implements engine.WarehouseDriver
func (b *BigQuery) WriteRecords(ctx context.Context, projectName string, streamName string, instanceID uuid.UUID, keys [][]byte, avros [][]byte, records []map[string]interface{}, timestamps []time.Time) error {
	// ensure all WriteRequest objects the same length
	if !(len(keys) == len(records) && len(keys) == len(avros) && len(keys) == len(timestamps)) {
		return fmt.Errorf("error: keys, avros, data, and timestamps do not all have the same length")
	}

	// create a BigQuery Row out of each of the records in the WriteRequest
	intRows := make([]*bq.StructSaver, len(keys))
	extRows := make([]*ExternalRow, len(keys))
	for i, key := range keys {
		// create insert ID
		insertIDBytes := append(key, timeutil.ToBytes(timestamps[i])...)
		insertID := base64.StdEncoding.EncodeToString(insertIDBytes)

		// create internal row
		intRows[i] = &bq.StructSaver{
			InsertID: insertID,
			Schema:   internalRowSchema,
			Struct: InternalRow{
				Key:       key,
				Data:      avros[i],
				Timestamp: timestamps[i],
			},
		}

		// create external row
		records[i]["__key"] = key
		records[i]["__timestamp"] = timestamps[i]
		extRows[i] = &ExternalRow{
			Data:     records[i],
			InsertID: insertID,
		}
	}

	// upload internals
	intU := b.Client.Dataset(internalDatasetName()).Table(internalTableName(instanceID)).Inserter()
	err := intU.Put(ctx, intRows)
	if err != nil {
		return err
	}

	// upload externals
	extU := b.Client.Dataset(externalDatasetName(projectName)).Table(externalTableName(streamName, instanceID)).Inserter()
	err = extU.Put(ctx, extRows)
	if err != nil {
		return err
	}

	// done
	return nil
}

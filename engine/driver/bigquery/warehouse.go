package bigquery

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"math"
	"math/big"
	"time"

	"cloud.google.com/go/bigquery"

	"gitlab.com/beneath-hq/beneath/engine/driver"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

const (
	expirationBuffer = 5 * time.Minute
)

// ExternalRow represents a record saved for external use (i.e. with columns matching schema)
// It implements bigquery.ValueSaver
type ExternalRow struct {
	Data     map[string]interface{}
	InsertID string
}

// Save implements bigquery.ValueSaver
func (r *ExternalRow) Save() (row map[string]bigquery.Value, insertID string, err error) {
	data := make(map[string]bigquery.Value, len(r.Data))
	for k, v := range r.Data {
		// we map []byte to hex normally, but mustn't do that for __key
		// (or else the hex encoding will be interpretted as base64 by BQ with terrible consequences)
		if k == "__key" {
			data[k] = base64.StdEncoding.EncodeToString(v.([]byte))
		} else {
			data[k] = r.recursiveSerialize(v)
		}
	}
	return data, r.InsertID, nil
}

// the bigquery client serializes every type correctly except big numbers and byte arrays;
// we handle those by a recursive search
// note: overrides in place
func (r *ExternalRow) recursiveSerialize(valT interface{}) bigquery.Value {
	switch val := valT.(type) {
	case float64:
		if math.IsNaN(val) {
			return "NaN"
		} else if math.IsInf(val, 1) {
			return "Infinity"
		} else if math.IsInf(val, -1) {
			return "-Infinity"
		}
		return val
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

// WriteToWarehouse implements beneath.WarehouseService
func (b BigQuery) WriteToWarehouse(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, rs []driver.Record) error {
	codec := s.GetCodec()

	// compute sensible timestamp at which to not even attempt the write
	expiration := time.Time{}
	if s.GetWarehouseRetention() != time.Duration(0) {
		expiration = time.Now().Add(-1 * s.GetWarehouseRetention()).Add(expirationBuffer)
	}

	// create a BigQuery Row out of each of the records
	expired := 0
	rows := make([]*ExternalRow, len(rs))
	for i, r := range rs {
		ts := r.GetTimestamp()
		if !expiration.IsZero() && ts.Before(expiration) {
			expired++
			continue
		}

		structured := r.GetStructured()
		structured["__timestamp"] = ts

		key, err := codec.MarshalKey(codec.PrimaryIndex, structured)
		if err != nil {
			return err
		}
		structured["__key"] = key

		insertIDBytes := append(key, timeutil.ToBytes(ts)...)
		insertID := base64.StdEncoding.EncodeToString(insertIDBytes)

		rows[i-expired] = &ExternalRow{
			Data:     structured,
			InsertID: insertID,
		}
	}

	// remove expired slots from records
	if expired != 0 {
		rows = rows[:len(rows)-expired]

		// there's a chance all records were expired, in which case we'll just stop here
		if len(rows) == 0 {
			return nil
		}
	}

	// save rows
	dataset := b.Client.Dataset(externalDatasetName(p))
	table := dataset.Table(externalTableName(s.GetStreamName(), i.GetStreamInstanceID()))
	err := table.Inserter().Put(ctx, rows)
	if err != nil {
		return err
	}

	return nil
}

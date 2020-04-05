package bigtable

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/bigtable"

	"gitlab.com/beneath-org/beneath/engine/driver"
	"gitlab.com/beneath-org/beneath/pkg/codec"
	"gitlab.com/beneath-org/beneath/pkg/codec/ext/tuple"
)

// Some of the row parsing logic and rowset generating logic should probably go in helper functions in schema.go.
// And I have a feeling some magic could be done here to make the functions shorter (e.g. LoadPrimaryIndexRange and LoadSecondaryIndexRange share a lot of logic).
// But this reads plainly top-to-bottom, and there's a lot of this code that's very-but-not-quite similar, so beware of messing up readability.

// Record encapsulates all possible information about a fetched record.
// It also implements driver.Record.
type Record struct {
	Offset       int64
	AvroData     []byte
	PrimaryKey   []byte
	SecondaryKey []byte
	Time         time.Time
	ProcessTime  time.Time
	Stream       driver.Stream
}

// IsNil implements driver.Record
func (r Record) IsNil() bool {
	return r.Offset == 0 && len(r.AvroData) == 0 && len(r.PrimaryKey) == 0 && len(r.SecondaryKey) == 0 && r.Time.IsZero() && r.ProcessTime.IsZero() && r.Stream == nil
}

// GetTimestamp implements driver.Record
func (r Record) GetTimestamp() time.Time {
	return r.Time
}

// GetAvro implements driver.Record
func (r Record) GetAvro() []byte {
	return r.AvroData
}

// GetStructured implements driver.Record
func (r Record) GetStructured() map[string]interface{} {
	structured, err := r.Stream.GetCodec().UnmarshalAvro(r.AvroData)
	if err != nil {
		panic(err)
	}

	structured, err = r.Stream.GetCodec().ConvertFromAvroNative(structured, false)
	if err != nil {
		panic(err)
	}

	return structured
}

// GetPrimaryKey implements driver.Record
func (r Record) GetPrimaryKey() []byte {
	return r.PrimaryKey
}

// LoadPrimaryIndexRange reads indexed records by primary key
func (b BigTable) LoadPrimaryIndexRange(ctx context.Context, s driver.Stream, i driver.StreamInstance, kr codec.KeyRange, limit int) ([]Record, codec.KeyRange, error) {
	// - Create bigtable range for primary index
	// - Load it
	// - If not normalized, return it
	// - Else, create bigtable rowset of offsets for log
	// - Load it
	// - Return it

	// prep
	instanceID := i.GetStreamInstanceID()
	normalized := s.GetCodec().PrimaryIndex.GetNormalize()
	hash := makeIndexHash(instanceID, s.GetCodec().PrimaryIndex.GetIndexID())

	// build rowset for primary index
	rs := makeIndexRowSetFromRange(hash, kr)

	// if not normalized, read directly
	if !normalized {
		var result []Record
		err := b.readIndexes(ctx, streamExpires(s), rs, limit, func(row bigtable.Row) bool {
			// parse key
			_, primaryKey := parseIndexKey(row.Key())

			// get avro column
			avroCol := row[indexesColumnFamilyName][0]
			if stripColumnFamily(avroCol.Column, indexesColumnFamilyName) != indexesAvroColumnName {
				panic(fmt.Errorf("unexpected column in LoadPrimaryIndexRange: %v", avroCol))
			}

			// create record
			record := Record{
				AvroData:   avroCol.Value,
				PrimaryKey: primaryKey,
				Time:       avroCol.Timestamp.Time(),
				Stream:     s,
			}

			// add record to result
			result = append(result, record)
			return true
		})
		if err != nil {
			return nil, codec.KeyRange{}, err
		}

		// done
		return result, nextPrimaryKeyRange(result, kr, limit), nil
	}

	// reading normalized data

	// get offsets from primary index
	var logRS bigtable.RowRangeList
	err := b.readIndexes(ctx, streamExpires(s), rs, limit, func(row bigtable.Row) bool {
		// get offset
		offsetCol := row[indexesColumnFamilyName][0]
		if stripColumnFamily(offsetCol.Column, indexesColumnFamilyName) != indexesOffsetColumnName {
			panic(fmt.Errorf("unexpected column in LoadPrimaryIndexRange: %v", offsetCol))
		}
		offset := bytesToInt(offsetCol.Value)

		// add to row range list
		rr := bigtable.PrefixRange(makeLogKey(instanceID, offset, nil))
		logRS = append(logRS, rr)
		return true
	})
	if err != nil {
		return nil, codec.KeyRange{}, err
	}

	// read offsets from log
	idx := 0
	result := make([]Record, len(logRS))
	err = b.readLog(ctx, streamExpires(s), logRS, 0, func(row bigtable.Row) bool {
		// parse key
		_, offset, primaryKey := parseLogKey(row.Key())

		// create record
		record := Record{
			Offset:     offset,
			PrimaryKey: primaryKey,
			Stream:     s,
		}

		// set record based on columns
		for _, column := range row[logColumnFamilyName] {
			switch stripColumnFamily(column.Column, logColumnFamilyName) {
			case logAvroColumnName:
				record.AvroData = column.Value
				record.Time = column.Timestamp.Time()
			case logProcessTimeColumnName:
				record.ProcessTime = bytesToTimeMs(column.Value)
			}
		}

		// add record to result
		result[idx] = record
		idx++
		return true
	})
	if err != nil {
		return nil, codec.KeyRange{}, err
	}

	// manually sort the result since log reads don't follow primary key order
	sort.Slice(result, func(i int, j int) bool {
		return bytes.Compare(result[i].PrimaryKey, result[j].PrimaryKey) < 0
	})

	return result, nextPrimaryKeyRange(result, kr, limit), nil
}

func nextPrimaryKeyRange(result []Record, prev codec.KeyRange, limit int) codec.KeyRange {
	var next codec.KeyRange
	if len(result) == limit {
		lastKey := result[len(result)-1].PrimaryKey
		nextBase := tuple.Successor(lastKey)
		if prev.Contains(nextBase) {
			next = codec.KeyRange{
				Base:     nextBase,
				RangeEnd: prev.RangeEnd,
			}
		}
	}
	return next
}

// LoadSecondaryIndexRange reads records indexed by secondary key
func (b BigTable) LoadSecondaryIndexRange(ctx context.Context, s driver.Stream, i driver.StreamInstance, index codec.Index, kr codec.KeyRange, limit int) ([]Record, codec.KeyRange, error) {
	// - Create bigtable range for secondary index
	// - Load it
	// - Put results into map[primary]Record
	// - Create bigtable rowset for primary index
	// - Load it
	// - Unite results with map and filter garbage secondary index keys out
	// - Delete garbage
	// - If primary index is normalized
	//   - Create bigtable rowset of offsets in log
	//   - Load it
	//   - Unite results with map
	// - Convert map to list, sort and return

	// prep
	c := s.GetCodec()
	instanceID := i.GetStreamInstanceID()
	primaryHash := makeIndexHash(instanceID, c.PrimaryIndex.GetIndexID())
	secondaryHash := makeIndexHash(instanceID, index.GetIndexID())

	// build rowset for secondary index
	rs := makeIndexRowSetFromRange(secondaryHash, kr)

	// load from secondary index
	var result map[string]Record
	err := b.readIndexes(ctx, streamExpires(s), rs, limit, func(row bigtable.Row) bool {
		// parse key
		_, secondaryKey := parseIndexKey(row.Key())

		// make primary key
		structured, err := c.UnmarshalKey(index, secondaryKey)
		if err != nil {
			panic(fmt.Errorf("secondary key %x doesn't match codec for %s: %v", secondaryKey, index.GetIndexID().String(), err.Error()))
		}
		primaryKey, err := c.MarshalKey(c.PrimaryIndex, structured)
		if err != nil {
			panic(fmt.Errorf("couldn't encode primary key with data from secondary key %x (index: %s): %s", secondaryKey, index.GetIndexID().String(), err.Error()))
		}

		// get offset column
		offsetCol := row[indexesColumnFamilyName][0]
		if offsetCol.Column != indexesOffsetColumnName {
			panic(fmt.Errorf("unexpected column in LoadSecondaryIndexRange: %v", offsetCol))
		}

		// store record
		result[byteSliceToString(primaryKey)] = Record{
			Offset:       bytesToInt(offsetCol.Value),
			PrimaryKey:   primaryKey,
			SecondaryKey: secondaryKey,
			Time:         offsetCol.Timestamp.Time(),
			Stream:       s,
		}

		return true
	})
	if err != nil {
		return nil, codec.KeyRange{}, err
	}

	// create rowset for lookup in primary index
	rl := make(bigtable.RowList, len(result))
	idx := 0
	for _, record := range result {
		rl[idx] = makeIndexKey(primaryHash, record.PrimaryKey)
		idx++
	}

	// load from primary index and consolidate
	var garbage []Record
	primaryNormalized := s.GetCodec().PrimaryIndex.GetNormalize()
	err = b.readIndexes(ctx, streamExpires(s), rl, 0, func(row bigtable.Row) bool {
		// parse key
		_, primaryKey := parseIndexKey(row.Key())
		primaryKeyString := byteSliceToString(primaryKey)

		// get record to update
		record := result[primaryKeyString]

		// update
		var primaryTime time.Time
		if primaryNormalized {
			// get offset
			offsetCol := row[indexesColumnFamilyName][0]
			if stripColumnFamily(offsetCol.Column, indexesColumnFamilyName) != indexesOffsetColumnName {
				panic(fmt.Errorf("unexpected column in loadFromPrimaryKeys: %v", offsetCol))
			}
			primaryTime = offsetCol.Timestamp.Time()
			record.Offset = bytesToInt(offsetCol.Value)
		} else {
			// get avro
			avroCol := row[indexesColumnFamilyName][0]
			if stripColumnFamily(avroCol.Column, indexesColumnFamilyName) != indexesAvroColumnName {
				panic(fmt.Errorf("unexpected column in loadFromPrimaryKeys: %v", avroCol))
			}
			primaryTime = avroCol.Timestamp.Time()
			record.AvroData = avroCol.Value
		}

		// if the timestamp doesn't match the one we loaded from the secondary index, it's garbage
		if record.Time.Before(primaryTime) {
			// add to garbage and remove from result
			garbage = append(garbage, record)
			delete(result, primaryKeyString)
		} else {
			// set updated record
			result[primaryKeyString] = record
		}

		return true
	})
	if err != nil {
		return nil, codec.KeyRange{}, err
	}

	// collect garbage from secondary index
	if len(garbage) > 0 {
		// create keys and deletes
		keys := make([]string, len(garbage))
		muts := make([]*bigtable.Mutation, len(garbage))
		for idx, record := range garbage {
			keys[idx] = makeIndexKey(secondaryHash, record.SecondaryKey)
			muts[idx] = makeSecondaryIndexDelete(record.Time.Add(time.Millisecond), index.GetNormalize())
		}

		// apply deletes
		_, indexesTable := b.tablesForStream(s)
		errs, err := indexesTable.ApplyBulk(ctx, keys, muts)
		if err != nil {
			return nil, codec.KeyRange{}, err
		} else if len(errs) != 0 {
			return nil, codec.KeyRange{}, errorFromApplyBulkErrors(errs)
		}
	}

	// if primary is normalized, get offsets
	if primaryNormalized {
		// create rowset for lookup in log
		idx := 0
		rrl := make(bigtable.RowRangeList, len(result))
		for _, record := range result {
			rrl[idx] = bigtable.PrefixRange(makeLogKey(instanceID, record.Offset, nil))
			idx++
		}

		// read and consolidate
		err := b.readLog(ctx, streamExpires(s), rrl, 0, func(row bigtable.Row) bool {
			// parse key
			_, offset, primaryKey := parseLogKey(row.Key())
			primaryKeyString := byteSliceToString(primaryKey)

			// get record to update
			record := result[primaryKeyString]
			record.Offset = offset

			// update record
			for _, column := range row[logColumnFamilyName] {
				switch stripColumnFamily(column.Column, logColumnFamilyName) {
				case logAvroColumnName:
					record.AvroData = column.Value
				case logProcessTimeColumnName:
					record.ProcessTime = bytesToTimeMs(column.Value)
				}
			}

			// write it back
			result[primaryKeyString] = record
			return true
		})
		if err != nil {
			return nil, codec.KeyRange{}, err
		}
	}

	// turn result into list
	idx = 0
	list := make([]Record, len(result))
	for _, record := range result {
		list[idx] = record
		idx++
	}

	// sort the list by secondary index
	sort.Slice(list, func(i int, j int) bool {
		return bytes.Compare(list[i].SecondaryKey, list[j].SecondaryKey) < 0
	})

	// compute next
	var next codec.KeyRange
	if len(list)+len(garbage) == limit {
		var endRecord Record
		if len(list) > 0 {
			endRecord = list[len(result)-1]
		}

		// we'll consider garbage collected rows too (should make the range narrower and next lookup possibly faster)
		if len(garbage) > 0 {
			for _, record := range garbage {
				if bytes.Compare(endRecord.SecondaryKey, record.SecondaryKey) < 0 {
					endRecord = record
				}
			}
		}

		next = codec.KeyRange{
			Base:     tuple.Successor(endRecord.SecondaryKey),
			RangeEnd: kr.RangeEnd,
		}
	}

	// done
	return list, next, nil
}

// LoadLogRange reads a range of offsets from the instance log
func (b BigTable) LoadLogRange(ctx context.Context, s driver.Stream, i driver.StreamInstance, from int64, to int64, limit int) ([]Record, error) {
	// prep
	instanceID := i.GetStreamInstanceID()

	// build rowset
	rs := makeLogRowSetFromRange(instanceID, from, to)

	// read and update records
	var result []Record
	err := b.readLog(ctx, streamExpires(s), rs, limit, func(row bigtable.Row) bool {
		// parse key
		_, offset, primaryKey := parseLogKey(row.Key())

		// make record
		record := Record{
			Offset:     offset,
			PrimaryKey: primaryKey,
			Stream:     s,
		}

		// set columns
		for _, column := range row[logColumnFamilyName] {
			switch stripColumnFamily(column.Column, logColumnFamilyName) {
			case logAvroColumnName:
				record.Time = column.Timestamp.Time()
				record.AvroData = column.Value
			case logProcessTimeColumnName:
				record.ProcessTime = bytesToTimeMs(column.Value)
			}
		}

		result = append(result, record)
		return true
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LoadExistingRecords loads existing rows that share a primary key with the input records
func (b BigTable) LoadExistingRecords(ctx context.Context, s driver.Stream, i driver.StreamInstance, records []driver.Record) (map[string]Record, error) {
	// - Create bigtable rowset of primary index keys
	// - Load it into a map[primary]Record
	// - If denormalized, return it
	// - Else, create rowset of offets into log
	// - Load it and consolidate it with map
	// - Return the map

	// prep
	c := s.GetCodec()
	instanceID := i.GetStreamInstanceID()
	normalized := c.PrimaryIndex.GetNormalize()
	hash := makeIndexHash(instanceID, c.PrimaryIndex.GetIndexID())

	// create rowset for lookup in primary index
	rl := make(bigtable.RowList, len(records))
	for idx, record := range records {
		key, err := c.MarshalKey(c.PrimaryIndex, record.GetStructured())
		if err != nil {
			return nil, err
		}
		rl[idx] = makeIndexKey(hash, key)
	}

	// load from primary index and consolidate
	var result map[string]Record
	err := b.readIndexes(ctx, streamExpires(s), rl, 0, func(row bigtable.Row) bool {
		// parse key
		_, primaryKey := parseIndexKey(row.Key())
		primaryKeyString := byteSliceToString(primaryKey)

		// prep record
		record := Record{
			PrimaryKey: primaryKey,
			Stream:     s,
		}

		// extract columns
		for _, col := range row[indexesColumnFamilyName] {
			switch stripColumnFamily(col.Column, indexesColumnFamilyName) {
			case indexesOffsetColumnName:
				record.Time = col.Timestamp.Time()
				record.Offset = bytesToInt(col.Value)
			case indexesAvroColumnName:
				record.Time = col.Timestamp.Time()
				record.AvroData = col.Value
			default:
				panic(fmt.Errorf("unexpected column in LoadExistingRecords: %v", col.Column))
			}
		}

		// save to result
		result[primaryKeyString] = record
		return true
	})
	if err != nil {
		return nil, err
	}

	// if not normalized, we're done
	if !normalized {
		return result, nil
	}

	// the records are normalized, so we have to fetch the avro data from the log

	// create rowset for lookup in log
	idx := 0
	rl = make(bigtable.RowList, len(result))
	for _, record := range result {
		rl[idx] = makeLogKey(instanceID, record.Offset, record.PrimaryKey)
		idx++
	}

	// read and consolidate
	err = b.readLog(ctx, streamExpires(s), rl, 0, func(row bigtable.Row) bool {
		// parse key
		_, _, primaryKey := parseLogKey(row.Key())
		primaryKeyString := byteSliceToString(primaryKey)

		// get record to update
		record := result[primaryKeyString]

		// update record
		for _, column := range row[logColumnFamilyName] {
			switch stripColumnFamily(column.Column, logColumnFamilyName) {
			case logAvroColumnName:
				record.AvroData = column.Value
			case logProcessTimeColumnName:
				record.ProcessTime = bytesToTimeMs(column.Value)
			}
		}

		// write it back
		result[primaryKeyString] = record
		return true
	})
	if err != nil {
		return nil, err
	}

	// done
	return result, nil
}

func errorFromApplyBulkErrors(errs []error) error {
	if len(errs) > 0 {
		return fmt.Errorf("got %d errors in ApplyBulk; first: %v", len(errs), errs[0])
	}
	return nil
}

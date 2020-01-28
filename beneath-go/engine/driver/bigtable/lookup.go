package bigtable

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/golang/protobuf/proto"

	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable/sequencer"
	pb "github.com/beneath-core/beneath-go/proto"
)

const (
	maxPartitions = 32
)

// ParseQuery implements driver.LookupService
func (b BigTable) ParseQuery(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, where queryparse.Query, compacted bool, partitions int) ([][]byte, [][]byte, error) {
	// checks not too many partitions
	if partitions > maxPartitions {
		return nil, nil, fmt.Errorf("the requested number of partitions exceeds the maximum number of supported partitions (%d)", partitions)
	}

	// check not filtering on un-compacted
	if !compacted {
		if !where.IsEmpty() {
			return nil, nil, fmt.Errorf("beneath currently doesn't support filters on non-compacted queries")
		}
	}

	// get instance log state
	state, err := b.Sequencer.GetState(ctx, makeSequencerKey(i.GetStreamInstanceID()))
	if err != nil {
		return nil, nil, err
	}

	if !compacted {
		return b.createUncompactedCursors(state, partitions)
	}

	// parse query
	index, kr, err := s.GetCodec().ParseQuery(where)
	if err != nil {
		return nil, nil, err
	}

	return b.createCompactedCursors(state, index, kr, partitions)
}

// creates replay and change cursors for a compacted query
func (b BigTable) createCompactedCursors(state sequencer.State, index codec.Index, kr codec.KeyRange, partitions int) ([][]byte, [][]byte, error) {
	// create replay cursor
	replay := &pb.Cursor{}
	replay.Type = pb.Cursor_INDEX
	replay.IndexId = int32(index.GetShortID())

	// if unique, lookup on key, else range
	if kr.IsSingle() {
		replay.IndexKey = kr.Base
	} else {
		replay.IndexStart = kr.Base
		replay.IndexEnd = kr.RangeEnd
	}

	// create change cursor
	changes := &pb.Cursor{}
	changes.Type = pb.Cursor_LOG
	changes.LogStart = state.NextStable

	// copy filter info over
	changes.IndexId = replay.IndexId
	changes.IndexKey = replay.IndexKey
	changes.IndexStart = replay.IndexStart
	changes.IndexEnd = replay.IndexEnd

	// compile and return
	replay1, err := compileCursor(replay)
	if err != nil {
		return nil, nil, err
	}

	changes1, err := compileCursor(changes)
	if err != nil {
		return nil, nil, err
	}

	return [][]byte{replay1}, [][]byte{changes1}, nil
}

func (b BigTable) createUncompactedCursors(state sequencer.State, partitions int) ([][]byte, [][]byte, error) {
	// to get the modulo arithmetic to match, NextStable must be greater than the number of partitions
	if state.NextStable < int64(partitions) {
		partitions = 1
	}

	// create replay cursors
	replays := make([][]byte, partitions)
	for i := 0; i < partitions; i++ {
		// configure
		replay := &pb.Cursor{}
		replay.Type = pb.Cursor_LOG
		replay.LogStart = int64(state.NextStable/int64(partitions)) * int64(i)
		if i+1 == partitions {
			replay.LogEnd = state.NextStable
		} else {
			replay.LogEnd = int64(state.NextStable/int64(partitions)) * int64(i+1)
		}

		// compile
		compiled, err := compileCursor(replay)
		if err != nil {
			return nil, nil, err
		}
		replays[i] = compiled
	}

	// for now, just a single change cursor, starting at NextStable
	change := &pb.Cursor{}
	change.Type = pb.Cursor_LOG
	change.LogStart = state.NextStable
	compiled, err := compileCursor(change)
	if err != nil {
		return nil, nil, err
	}
	changes := [][]byte{compiled}

	// done
	return replays, changes, nil
}

func compileCursor(cursor *pb.Cursor) ([]byte, error) {
	compiled, err := proto.Marshal(cursor)
	return compiled, err
}

func decompileCursor(bs []byte) (*pb.Cursor, error) {
	cursor := &pb.Cursor{}
	err := proto.Unmarshal(bs, cursor)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

// ReadCursor implements driver.LookupService
func (b BigTable) ReadCursor(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursorRaw []byte, limit int) (driver.RecordsIterator, error) {
	// parse cursor
	_, err := decompileCursor(cursorRaw)
	if err != nil {
		return nil, err
	}

	// create row set
	// use lookupIndexedRecords or lookupLogRecords

	panic("implement")
}

// WriteRecords implements driver.LookupService
// NOTE/TODO: This implementation suffers from a race condition. It may arise when two concurrent calls to WriteRecords
// write to the same primary key, and both modify the secondary index in different ways. This leads to garbage in the
// secondary index. We solve the problem by disregarding normalization options and forcing ReadCursor to lookup secondary
// index values in the primary index, so that it can compare timestamps and pickup garbage. But ideally, secondary indexes
// could be denormalized and not require multiple lookups.
func (b BigTable) WriteRecords(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, records []driver.Record) error {
	// Prepare
	codec := s.GetCodec()
	retention := s.GetRetention()
	expires := streamExpires(s)
	instanceID := i.GetStreamInstanceID()
	processTime := time.Now()

	// get existing values (to delete secondary indexes)
	var existingToReplace map[string]internalRecord
	if len(codec.SecondaryIndexes) > 0 {
		existing, err := b.lookupExistingRecords(ctx, s, i, records)
		if err != nil {
			return err
		}
		existingToReplace = existing
	}

	// get offset
	batch, err := b.Sequencer.BeginBatch(ctx, makeSequencerKey(i.GetStreamInstanceID()), len(records))
	if err != nil {
		return err
	}

	// prepare mutations
	writeLogKeys := make([]string, len(records))
	writeLogMuts := make([]*bigtable.Mutation, len(records))

	var writePrimaryKeys []string
	var writePrimaryMuts []*bigtable.Mutation

	var writeSecondaryKeys []string
	var writeSecondaryMuts []*bigtable.Mutation

	var deleteSecondaryKeys []string
	var deleteSecondaryMuts []*bigtable.Mutation

	// create mutations
	for idx, record := range records {
		// make primary key
		structured := record.GetStructured()
		primaryKey, err := codec.MarshalKey(codec.PrimaryIndex, structured)
		if err != nil {
			return err
		}

		// offset
		offset := batch.NumberAtOffset(idx)

		// timestamp (if retention is configured, set rowTime to expiration timestamp instead)
		rowTime := record.GetTimestamp()
		if expires {
			rowTime = rowTime.Add(retention)
		}

		// create log append mutation
		writeLogMuts[idx] = makeLogInsert(rowTime, processTime, record.GetAvro())
		writeLogKeys[idx] = makeLogKey(instanceID, offset, primaryKey)

		// if earlier than existing record (with same primary key), no need to handle
		existing := existingToReplace[byteSliceToString(primaryKey)]
		if !existing.IsNil() {
			// skip updating primary/secondary indexes if existing has later timestamp
			if record.GetTimestamp().Before(existing.GetTimestamp()) {
				continue
			}
		}

		// create primary key write
		norm := codec.PrimaryIndex.GetNormalize()
		writePrimaryMuts = append(writePrimaryMuts, makePrimaryIndexInsert(rowTime, norm, offset, record.GetAvro()))
		writePrimaryKeys = append(writePrimaryKeys, makePrimaryIndexKey(instanceID, primaryKey))

		// create secondary indexes
		for _, secondaryIndex := range codec.SecondaryIndexes {
			// compute new secondary index key
			newSecondaryKey, err := codec.MarshalKey(secondaryIndex, record.GetStructured())
			if err != nil {
				return err
			}

			// add secondary index write
			norm := secondaryIndex.GetNormalize()
			writeSecondaryMuts = append(writeSecondaryMuts, makeSecondaryIndexInsert(rowTime, norm, offset, record.GetAvro()))
			writeSecondaryKeys = append(writeSecondaryKeys, makeSecondaryIndexKey(instanceID, newSecondaryKey, primaryKey))

			// if new secondary index key is different from the existing, delete the existing
			if !existing.IsNil() {
				// compute existing secondary index key
				existingSecondaryKey, err := codec.MarshalKey(secondaryIndex, existing.GetStructured())
				if err != nil {
					return err
				}

				// separately delete existing secondary index entry if the secondary index keys are different
				if !bytes.Equal(newSecondaryKey, existingSecondaryKey) {
					deleteSecondaryMuts = append(deleteSecondaryMuts, makeSecondaryIndexDelete(rowTime, norm))
					deleteSecondaryKeys = append(deleteSecondaryKeys, makeSecondaryIndexKey(instanceID, existingSecondaryKey, primaryKey))
				}
			}
		}
	}

	// Execute mutations in this order
	//   delete current secondary keys
	//   save to log
	//   save to primary
	//   save to secondary

	logTable, indexesTable := b.tablesForStream(s)

	errs, err := indexesTable.ApplyBulk(ctx, deleteSecondaryKeys, deleteSecondaryMuts)
	if err != nil {
		return err
	} else if len(errs) != 0 {
		return errorFromApplyBulkErrors(errs)
	}

	errs, err = logTable.ApplyBulk(ctx, writeLogKeys, writeLogMuts)
	if err != nil {
		return err
	} else if len(errs) != 0 {
		return errorFromApplyBulkErrors(errs)
	}

	errs, err = indexesTable.ApplyBulk(ctx, writePrimaryKeys, writePrimaryMuts)
	if err != nil {
		return err
	} else if len(errs) != 0 {
		return errorFromApplyBulkErrors(errs)
	}

	errs, err = indexesTable.ApplyBulk(ctx, writeSecondaryKeys, writeSecondaryMuts)
	if err != nil {
		return err
	} else if len(errs) != 0 {
		return errorFromApplyBulkErrors(errs)
	}

	// commit offset
	err = b.Sequencer.CommitBatch(ctx, batch)
	if err != nil {
		// probably a timeout error
		// maybe clean up written data to avoid duplicates in "log"?
		return err
	}

	return nil
}

func (b BigTable) lookupExistingRecords(ctx context.Context, s driver.Stream, i driver.StreamInstance, records []driver.Record) (map[string]internalRecord, error) {
	// create internal records for lookup by primary key
	codec := s.GetCodec()
	var result map[string]internalRecord
	for _, record := range records {
		key, err := codec.MarshalKey(codec.PrimaryIndex, record.GetStructured())
		if err != nil {
			return nil, err
		}
		keyStr := byteSliceToString(key)
		result[keyStr] = internalRecord{
			PrimaryKey: key,
		}
	}

	// execute load by primary keys
	err := b.loadFromPrimaryKeys(ctx, s, i, result)
	if err != nil {
		return nil, err
	}

	// remove the ones not found
	for key, val := range result {
		// checking on Timestamp because we can rely on loadFromPrimaryKeys setting it
		if val.Timestamp == 0 {
			delete(result, key)
		}
	}

	// done
	return result, nil
}

func (b BigTable) loadFromOffsets(ctx context.Context, s driver.Stream, i driver.StreamInstance, records map[string]internalRecord) error {
	// build rowset
	idx := 0
	rrl := make(bigtable.RowRangeList, len(records))
	for _, record := range records {
		rrl[idx] = bigtable.PrefixRange(makeLogKeyPrefix(i.GetStreamInstanceID(), record.Offset)) // can't rely on record.PrimaryKey being set
		idx++
	}

	// read and update records
	err := b.readLog(ctx, streamExpires(s), rrl, func(row bigtable.Row) bool {
		// parse key
		_, offset, primaryKey := parseLogKey(row.Key())
		primaryKeyString := byteSliceToString(primaryKey)

		// get record to update
		record := records[primaryKeyString]
		record.Offset = offset

		// update record
		for _, column := range row[logColumnFamilyName] {
			if column.Column == logAvroColumnName {
				record.AvroData = column.Value
				record.Timestamp = column.Timestamp
				break
			}
		}
		records[primaryKeyString] = record

		return true
	})
	if err != nil {
		return err
	}

	// consolidate iteratively
	return nil
}

func (b BigTable) loadFromPrimaryKeys(ctx context.Context, s driver.Stream, i driver.StreamInstance, records map[string]internalRecord) error {
	// prep
	normalized := s.GetCodec().PrimaryIndex.GetNormalize()

	// build rowset
	idx := 0
	rl := make(bigtable.RowList, len(records))
	for key := range records {
		rl[idx] = makePrimaryIndexKey(i.GetStreamInstanceID(), []byte(key))
		idx++
	}

	// read and consolidate
	err := b.readIndexes(ctx, streamExpires(s), rl, func(row bigtable.Row) bool {
		// parse key
		_, primaryKey := parsePrimaryIndexKey(row.Key())
		primaryKeyString := byteSliceToString(primaryKey)

		// get record to update
		record := records[primaryKeyString]

		// update
		var ts bigtable.Timestamp
		if normalized {
			// get offset
			offsetCol := row[indexesColumnFamilyName][0]
			if offsetCol.Column != indexesOffsetColumnName {
				panic(fmt.Errorf("unexpected column in loadFromPrimaryKeys: %v", offsetCol))
			}
			ts = offsetCol.Timestamp
			record.Offset = bytesToInt(offsetCol.Value)
		} else {
			// get avro
			avroCol := row[indexesColumnFamilyName][0]
			if avroCol.Column != indexesAvroColumnName {
				panic(fmt.Errorf("unexpected column in loadFromPrimaryKeys: %v", avroCol))
			}
			ts = avroCol.Timestamp
			record.AvroData = avroCol.Value
		}

		// check timestamp
		if len(record.SecondaryKey) != 0 && record.Timestamp < ts {
			// TODO
			// secondary index miss
		}

		// set updated record
		record.Timestamp = ts
		records[primaryKeyString] = record
		return true
	})
	if err != nil {
		return err
	}

	return nil
}

func (b BigTable) loadFromSecondaryKeys(ctx context.Context, s driver.Stream, i driver.StreamInstance, secondaryKeys [][]byte) (map[string]internalRecord, error) {
	// build rowset
	rrl := make(bigtable.RowRangeList, len(secondaryKeys))
	for idx, secondaryKey := range secondaryKeys {
		rrl[idx] = bigtable.PrefixRange(makeSecondaryIndexKeyPrefix(i.GetStreamInstanceID(), secondaryKey))
	}

	// read secondary keys into internalRecords (indexed by primary key)
	var result map[string]internalRecord
	err := b.readIndexes(ctx, streamExpires(s), rrl, func(row bigtable.Row) bool {
		//

		// parse key
		_, secondaryKey, primaryKey := parseSecondaryIndexKey(row.Key())

		// get column

		offsetCol := row[indexesColumnFamilyName][0]
		if offsetCol.Column != indexesOffsetColumnName {
			panic(fmt.Errorf("unexpected column in loadFromSecondaryKeys: %v", offsetCol))
		}

		// store record
		result[byteSliceToString(primaryKey)] = internalRecord{
			Offset:       bytesToInt(offsetCol.Value),
			PrimaryKey:   primaryKey,
			SecondaryKey: secondaryKey,
			Timestamp:    offsetCol.Timestamp,
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func errorFromApplyBulkErrors(errs []error) error {
	if len(errs) > 0 {
		return fmt.Errorf("got %d errors in ApplyBulk; first: %v", len(errs), errs[0])
	}
	return nil
}

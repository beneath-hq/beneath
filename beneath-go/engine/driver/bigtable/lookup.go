package bigtable

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/golang/protobuf/proto"

	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/core/mathutil"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/engine/driver"
	pb "github.com/beneath-core/beneath-go/engine/driver/bigtable/proto"
	"github.com/beneath-core/beneath-go/engine/driver/bigtable/sequencer"
)

const (
	maxPartitions   = 32
	maxCursors      = 32
	maxLimit        = 1000
	maxLoadsPerRead = 5
)

var (
	// ErrCorruptCursor is returned from ReadCursor for cursors that have been altered
	ErrCorruptCursor = errors.New("malformed or corrupted cursor")

	// ErrInvalidLimit is returned from ReadCursor for limits outside allowed bounds
	ErrInvalidLimit = errors.New("limit must be greater than 0 and lower than 1000")
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

	// create cursors
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

// creates replay and change cursors for an uncompacted (pure log) query
func (b BigTable) createUncompactedCursors(state sequencer.State, partitions int) ([][]byte, [][]byte, error) {
	// to get the modulo arithmetic to match, NextStable must be greater than the number of partitions
	if state.NextStable < int64(partitions) {
		partitions = 1
	}

	// create replay cursors
	replays := make([][]byte, partitions)
	for i := 0; i < partitions; i++ {
		// compute log segment start and end
		var start, end int64
		start = int64(state.NextStable/int64(partitions)) * int64(i)
		end = int64(state.NextStable/int64(partitions)) * int64(i+1)
		if i+1 == partitions {
			end = state.NextStable
		}

		// compile
		cursor := &pb.Cursor{
			Type:     pb.Cursor_LOG,
			LogStart: start,
			LogEnd:   end,
		}
		set, err := compileCursor(cursor)
		if err != nil {
			return nil, nil, err
		}
		replays[i] = set
	}

	// for now, just a single change cursor, starting at NextStable
	cursor := &pb.Cursor{
		Type:     pb.Cursor_LOG,
		LogStart: state.NextStable,
	}
	set, err := compileCursor(cursor)
	if err != nil {
		return nil, nil, err
	}
	changes := [][]byte{set}

	// done
	return replays, changes, nil
}

// creates replay and change cursors for a compacted query
func (b BigTable) createCompactedCursors(state sequencer.State, index codec.Index, kr codec.KeyRange, partitions int) ([][]byte, [][]byte, error) {
	// create replay cursor
	replay := &pb.Cursor{}
	replay.Type = pb.Cursor_INDEX

	replay.IndexId = int32(index.GetShortID())
	replay.IndexStart = kr.Base
	replay.IndexEnd = kr.RangeEnd

	// create change cursor
	changes := &pb.Cursor{}
	changes.Type = pb.Cursor_LOG
	changes.LogStart = state.NextStable

	// copy filter info over
	changes.IndexId = replay.IndexId
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

func compileCursor(cursor *pb.Cursor) ([]byte, error) {
	return compileCursorSet(&pb.CursorSet{
		Cursors: []*pb.Cursor{cursor},
	})
}

func compileCursorSet(set *pb.CursorSet) ([]byte, error) {
	compiled, err := proto.Marshal(set)
	return compiled, err
}

func decompileCursorSet(bs []byte) (*pb.CursorSet, error) {
	set := &pb.CursorSet{}
	err := proto.Unmarshal(bs, set)
	if err != nil {
		return nil, err
	}
	return set, nil
}

// ReadCursor implements driver.LookupService
func (b BigTable) ReadCursor(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursorSet []byte, limit int) (driver.RecordsIterator, error) {
	// check limit
	if limit == 0 || limit > maxLimit {
		return nil, ErrInvalidLimit
	}

	// parse cursor set
	set, err := decompileCursorSet(cursorSet)
	if err != nil {
		return nil, err
	}

	// check lengths
	if len(set.Cursors) == 0 {
		return nil, ErrCorruptCursor
	} else if len(set.Cursors) > maxCursors {
		return nil, ErrCorruptCursor
	}

	// check all same type
	first := set.Cursors[0]
	for _, cursor := range set.Cursors[1:] {
		if cursor.Type != first.Type {
			return nil, ErrCorruptCursor
		}
	}

	// if index scan, check just one cursor
	if first.Type == pb.Cursor_INDEX {
		if len(set.Cursors) != 1 {
			return nil, ErrCorruptCursor
		}
	}

	// parse based on type
	switch first.Type {
	case pb.Cursor_LOG:
		return b.readLogCursors(ctx, s, i, set, limit)
	case pb.Cursor_INDEX:
		return b.readIndexCursor(ctx, s, i, first, limit)
	default:
		return nil, ErrCorruptCursor
	}
}

// recordsIterator implements driver.RecordsIterator
type recordsIterator struct {
	idx        int
	records    []Record
	nextCursor []byte
}

// Next implements driver.RecordsIterator
func (i recordsIterator) Next() driver.Record {
	if len(i.records) == i.idx {
		return nil
	}
	r := i.records[i.idx]
	i.idx++
	return r
}

// NextCursor implements driver.RecordsIterator
func (i recordsIterator) NextCursor() []byte {
	return i.nextCursor
}

// batchesIterator implements driver.RecordsIterator
type batchesIterator struct {
	recordIdx  int
	batches    [][]Record
	nextCursor []byte
}

// Next implements driver.RecordsIterator
func (i batchesIterator) Next() driver.Record {
	if len(i.batches) == 0 {
		return nil
	}
	batch := i.batches[0]
	r := batch[i.recordIdx]
	i.recordIdx++
	if len(batch) == i.recordIdx {
		i.batches = i.batches[1:]
		i.recordIdx = 0
	}
	return r
}

// NextCursor implements driver.RecordsIterator
func (i batchesIterator) NextCursor() []byte {
	return i.nextCursor
}

func (b BigTable) readLogCursors(ctx context.Context, s driver.Stream, i driver.StreamInstance, set *pb.CursorSet, limit int) (driver.RecordsIterator, error) {
	// important log invariants:
	// (1) for two records r1 and r2 where offset(r2) > offset(r1) we have that processTime(r2) >= processTime(r1)
	// (2) for a cursor where LogEnd != 0, we know that offset=LogEnd is already written or is expired

	// important loop invariants:
	// - for up to n iterations, cursors[0] is the next cursor we should fetch
	// - newly generated cursors are _appended_ to cursors
	// - len(cursors) must stay <= maxCursors

	cursors := set.Cursors
	var results [][]Record

	n := mathutil.MinInt(len(cursors), maxLoadsPerRead)
	for counter := 0; counter < n; counter++ {
		// get cursor
		cursor := cursors[0]
		cursors = cursors[1:]

		// check range is valid
		if cursor.LogStart < 0 || cursor.LogEnd < 0 {
			return nil, ErrCorruptCursor
		}
		if cursor.LogEnd != 0 && (cursor.LogStart >= cursor.LogEnd) {
			return nil, ErrCorruptCursor
		}

		// compute limit and end, such that we get one row of the next segment if there's a segment missing in the log
		lim := limit
		end := cursor.LogEnd
		if cursor.LogEnd != 0 {
			lim = mathutil.MinInt(lim, int(cursor.LogEnd-cursor.LogStart))
			end = cursor.LogEnd + 1
		}

		// read records
		records, err := b.LoadLogRange(ctx, s, i, cursor.LogStart, end, lim)
		if err != nil {
			return nil, err
		}

		// if records is empty, we're either a) at the tip of an infinite cursor, or b) the log segment is timed out by invariant (2) (because LogEnd+1 wasn't returned)
		if len(records) == 0 {
			if cursor.LogEnd == 0 {
				// we're at the tip of an infinite cursor, so add it back to cursors
				cursors = append(cursors, cursor)
			}
			// else the cursor covers a timed out segment, so we skip it
			continue
		}

		// check for gaps in records
		nextOffset := cursor.LogStart // offset we expect of the next record in the iteration
		for idx, record := range records {
			// easy case: we found the row we expected
			if record.Offset == nextOffset {
				nextOffset++
				continue
			}

			// we now know that record.Offset != nextOffset (so there's a gap)

			// determine if gap is timed out (remember: record.Time is the event time, record.ProcessTime is the bigtable time of the batch -- see invariant (1))
			gapTimedOut := time.Since(record.ProcessTime) >= (b.Sequencer.TTL + b.Sequencer.MaxDrift)

			// skip the gap if it's timed out
			if gapTimedOut {
				nextOffset = record.Offset + 1
				continue
			}

			// gap hasn't timed out

			// we leave room for one cursor to be added at the end of the loop
			if len(cursors)+1 == maxCursors {
				records = records[:idx]
				break
			}

			// add gap to cursors
			cursors = append(cursors, &pb.Cursor{
				Type:       pb.Cursor_LOG,
				LogStart:   nextOffset,
				LogEnd:     record.Offset,
				IndexId:    cursor.IndexId,
				IndexStart: cursor.IndexStart,
				IndexEnd:   cursor.IndexEnd,
			})

			// update nextOffset to reflect the gap we just skipped
			nextOffset = record.Offset + 1
		}

		// if this cursor spans more records than we fetched, add a cursor that spans the remains
		if cursor.LogEnd == 0 || cursor.LogEnd > nextOffset {
			cursors = append(cursors, &pb.Cursor{
				Type:       pb.Cursor_LOG,
				LogStart:   nextOffset,
				LogEnd:     cursor.LogEnd,
				IndexId:    cursor.IndexId,
				IndexStart: cursor.IndexStart,
				IndexEnd:   cursor.IndexEnd,
			})
		}

		// records might end with offset (cursor.LogEnd + 1) if there were missing rows
		// remove it if that's the case
		if len(records) > 0 && cursor.LogEnd > 0 {
			lastIdx := len(records) - 1
			if records[lastIdx].Offset == cursor.LogEnd {
				records = records[:lastIdx]
			}
		}

		// add records to accumulator
		if len(records) > 0 {
			results = append(results, records)
		}
	}

	// sort cursors
	sort.Slice(cursors, func(i int, j int) bool {
		return cursors[i].LogStart < cursors[j].LogStart
	})

	// compile cursors
	var nextCursor []byte
	if len(cursors) > 0 {
		compiled, err := compileCursorSet(&pb.CursorSet{Cursors: cursors})
		if err != nil {
			return nil, err
		}
		nextCursor = compiled
	}

	return batchesIterator{
		batches:    results,
		nextCursor: nextCursor,
	}, nil
}

func (b BigTable) readIndexCursor(ctx context.Context, s driver.Stream, i driver.StreamInstance, cursor *pb.Cursor, limit int) (driver.RecordsIterator, error) {
	// create key range from cursor
	kr := codec.KeyRange{
		Base:     cursor.IndexStart,
		RangeEnd: cursor.IndexEnd,
	}

	// load from either primary or secondary index
	var records []Record
	var next codec.KeyRange
	var err error
	if cursor.IndexId == 0 {
		// load from primary index
		records, next, err = b.LoadPrimaryIndexRange(ctx, s, i, kr, limit)
	} else {
		// find secondary index
		secondaryIndex := s.GetCodec().FindIndexByShortID(int(cursor.IndexId))
		if secondaryIndex == nil {
			return nil, ErrCorruptCursor
		}

		// load
		records, next, err = b.LoadSecondaryIndexRange(ctx, s, i, secondaryIndex, kr, limit)
	}

	// check error
	if err != nil {
		return nil, err
	}

	// encode new cursor
	var nextCursor []byte
	if !next.IsNil() {
		cursor.IndexStart = next.Base
		cursor.IndexEnd = next.RangeEnd
		compiled, err := compileCursor(cursor)
		if err != nil {
			return nil, err
		}
		nextCursor = compiled
	}

	// return iterator
	return recordsIterator{
		records:    records,
		nextCursor: nextCursor,
	}, nil
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
	normalizePrimary := codec.PrimaryIndex.GetNormalize()
	primaryHash := makeIndexHash(instanceID, codec.PrimaryIndex.GetIndexID())

	// get existing values (to delete secondary indexes)
	var existingToReplace map[string]Record
	if len(codec.SecondaryIndexes) > 0 {
		existing, err := b.LoadExistingRecords(ctx, s, i, records)
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
		writeLogMuts[idx] = makeLogInsert(rowTime, batch.BigtableTime, record.GetAvro())
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
		writePrimaryMuts = append(writePrimaryMuts, makePrimaryIndexInsert(rowTime, normalizePrimary, offset, record.GetAvro()))
		writePrimaryKeys = append(writePrimaryKeys, makeIndexKey(primaryHash, primaryKey))

		// create secondary indexes
		for _, secondaryIndex := range codec.SecondaryIndexes {
			// compute new secondary index key
			newSecondaryKey, err := codec.MarshalKey(secondaryIndex, record.GetStructured())
			if err != nil {
				return err
			}

			// prep
			normalizeSecondary := secondaryIndex.GetNormalize()
			secondaryHash := makeIndexHash(instanceID, secondaryIndex.GetIndexID())

			// add secondary index write
			writeSecondaryMuts = append(writeSecondaryMuts, makeSecondaryIndexInsert(rowTime, normalizeSecondary, offset, record.GetAvro()))
			writeSecondaryKeys = append(writeSecondaryKeys, makeIndexKey(secondaryHash, newSecondaryKey))

			// if new secondary index key is different from the existing, delete the existing
			if !existing.IsNil() {
				// compute existing secondary index key
				existingSecondaryKey, err := codec.MarshalKey(secondaryIndex, existing.GetStructured())
				if err != nil {
					return err
				}

				// separately delete existing secondary index entry if the secondary index keys are different
				if !bytes.Equal(newSecondaryKey, existingSecondaryKey) {
					deleteSecondaryMuts = append(deleteSecondaryMuts, makeSecondaryIndexDelete(rowTime, normalizeSecondary))
					deleteSecondaryKeys = append(deleteSecondaryKeys, makeIndexKey(secondaryHash, existingSecondaryKey))
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

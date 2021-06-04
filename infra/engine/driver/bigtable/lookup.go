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

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/driver/bigtable/proto"
	"github.com/beneath-hq/beneath/infra/engine/driver/bigtable/sequencer"
	"github.com/beneath-hq/beneath/pkg/codec"
	"github.com/beneath-hq/beneath/pkg/mathutil"
	"github.com/beneath-hq/beneath/pkg/queryparse"
)

const (
	maxPartitions    = 32
	maxCursors       = 32
	maxLimit         = 1000
	maxLoadsPerRead  = 5
	expirationBuffer = 5 * time.Minute
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

	if compacted {
		if !s.GetUseIndex() {
			return nil, nil, fmt.Errorf("cannot execute indexed queries on stream with useIndex=false")
		}
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
	index, kr, err := s.GetCodec().ParseIndexQuery(where)
	if err != nil {
		return nil, nil, err
	}

	return b.createCompactedCursors(state, index, kr, partitions)
}

// creates replay and change cursors for an uncompacted (pure log) query
func (b BigTable) createUncompactedCursors(state sequencer.State, partitions int) ([][]byte, [][]byte, error) {
	// to get the modulo arithmetic to match, NextStable must be greater than the number of partitions
	if state.NextStable < int64(partitions) {
		partitions = int(state.NextStable)
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

		// skip if start == end (e.g. if NextStable is 0)
		if start == end {
			continue
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

// Peek implements driver.LookupService
func (b BigTable) Peek(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) ([]byte, []byte, error) {
	// get instance log state
	state, err := b.Sequencer.GetState(ctx, makeSequencerKey(i.GetStreamInstanceID()))
	if err != nil {
		return nil, nil, err
	}

	// create rewind cursor
	cursor := &pb.Cursor{
		Type:   pb.Cursor_PEEK,
		LogEnd: state.NextStable,
	}
	rewind, err := compileCursor(cursor)
	if err != nil {
		return nil, nil, err
	}

	// create changes cursor
	cursor = &pb.Cursor{
		Type:     pb.Cursor_LOG,
		LogStart: state.NextStable,
	}
	changes, err := compileCursor(cursor)
	if err != nil {
		return nil, nil, err
	}

	// done
	return rewind, changes, nil
}

// ReadCursor implements driver.LookupService
func (b BigTable) ReadCursor(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance, cursorSet []byte, limit int) (driver.RecordsIterator, error) {
	// check limit
	if limit == 0 || limit > maxLimit {
		return nil, ErrInvalidLimit
	}

	// check cursor
	if cursorSet == nil {
		return &recordsIterator{}, nil
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

	// if not log scan, check just one cursor
	if first.Type != pb.Cursor_LOG {
		if len(set.Cursors) != 1 {
			return nil, ErrCorruptCursor
		}
	}

	// check cursor type
	if first.Type == pb.Cursor_INDEX {
		if !s.GetUseIndex() {
			return nil, fmt.Errorf("Can't read index cursor on stream with useIndex=false (unexpected error, how did you get this cursor?)")
		}
	}

	// parse based on type
	switch first.Type {
	case pb.Cursor_PEEK:
		return b.readPeekCursor(ctx, s, i, first, limit)
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
	reverse    bool
	records    []Record
	nextCursor []byte
}

// Next implements driver.RecordsIterator
func (i *recordsIterator) Next() bool {
	i.idx++
	return len(i.records) >= i.idx
}

// Record implements driver.RecordsIterator
func (i *recordsIterator) Record() driver.Record {
	if i.idx == 0 || i.idx > len(i.records) {
		panic("invalid call to Record")
	}
	if i.reverse {
		return i.records[len(i.records)-i.idx]
	}
	return i.records[i.idx-1]
}

// NextCursor implements driver.RecordsIterator
func (i *recordsIterator) NextCursor() []byte {
	return i.nextCursor
}

// batchesIterator implements driver.RecordsIterator
type batchesIterator struct {
	recordIdx  int
	batches    [][]Record
	next       Record
	nextCursor []byte
}

// Next implements driver.RecordsIterator
func (i *batchesIterator) Next() bool {
	if len(i.batches) == 0 {
		return false
	}

	batch := i.batches[0]
	r := batch[i.recordIdx]
	i.recordIdx++
	if len(batch) == i.recordIdx {
		i.batches = i.batches[1:]
		i.recordIdx = 0
	}

	i.next = r
	return true
}

// Record implements driver.RecordsIterator
func (i *batchesIterator) Record() driver.Record {
	return i.next
}

// NextCursor implements driver.RecordsIterator
func (i *batchesIterator) NextCursor() []byte {
	return i.nextCursor
}

func (b BigTable) readPeekCursor(ctx context.Context, s driver.Stream, i driver.StreamInstance, cursor *pb.Cursor, limit int) (driver.RecordsIterator, error) {
	// prepare range
	end := cursor.LogEnd
	start := mathutil.MaxInt64(0, end-int64(limit))

	// load from log
	records, err := b.LoadLogRange(ctx, s, i, start, end, limit)

	// check error
	if err != nil {
		return nil, err
	}

	// encode new cursor
	var nextCursor []byte
	if start != 0 {
		cursor.LogEnd = start
		compiled, err := compileCursor(cursor)
		if err != nil {
			return nil, err
		}
		nextCursor = compiled
	}

	// return iterator
	return &recordsIterator{
		reverse:    true,
		records:    records,
		nextCursor: nextCursor,
	}, nil
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

	return &batchesIterator{
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
	return &recordsIterator{
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
	logRetention := s.GetLogRetention()
	indexRetention := s.GetIndexRetention()
	logExpires := logExpires(s)
	indexExpires := indexExpires(s)
	useIndex := s.GetUseIndex()
	instanceID := i.GetStreamInstanceID()
	normalizePrimary := codec.PrimaryIndex.GetNormalize()
	primaryHash := makeIndexHash(instanceID, codec.PrimaryIndex.GetIndexID())
	nowForExpirationChecks := toPersistedTime(time.Now(), expirationBuffer)

	// get existing values (to delete secondary indexes)
	var existingToReplace map[string]Record
	if useIndex && len(codec.SecondaryIndexes) > 0 {
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
	var writeLogKeys []string
	var writeLogMuts []*bigtable.Mutation

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

		// log row timestamp (if retention is configured, set rowTime to expiration timestamp instead)
		logRowTime := record.GetTimestamp()
		if logExpires {
			logRowTime = toPersistedTime(logRowTime, logRetention)
		}

		// write unless row is already practically expired
		if !logExpires || logRowTime.After(nowForExpirationChecks) {
			// create log append mutation
			writeLogMuts = append(writeLogMuts, makeLogInsert(logRowTime, batch.BigtableTime, record.GetAvro()))
			writeLogKeys = append(writeLogKeys, makeLogKey(instanceID, offset, primaryKey))
		}

		// if not writing index, we can continue
		if !useIndex {
			continue
		}

		// index row timestamp (if retention is configured, rowTime is set to expiration timestamp instead)
		indexRowTime := record.GetTimestamp()
		if indexExpires {
			indexRowTime = toPersistedTime(indexRowTime, indexRetention)
		}

		// if row is practically already expired, don't write it
		if indexExpires && indexRowTime.Before(nowForExpirationChecks) {
			continue
		}

		// if earlier than existing record (with same primary key), no need to handle
		existing := existingToReplace[byteSliceToString(primaryKey)]
		if !existing.IsNil() {
			// skip updating primary/secondary indexes if existing has later timestamp
			if record.GetTimestamp().Before(existing.GetTimestamp()) {
				continue
			}
		}

		// create primary key write
		writePrimaryMuts = append(writePrimaryMuts, makePrimaryIndexInsert(indexRowTime, normalizePrimary, offset, record.GetAvro()))
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
			writeSecondaryMuts = append(writeSecondaryMuts, makeSecondaryIndexInsert(indexRowTime, normalizeSecondary, offset, record.GetAvro()))
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
					deleteSecondaryMuts = append(deleteSecondaryMuts, makeSecondaryIndexDelete(indexRowTime, normalizeSecondary))
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

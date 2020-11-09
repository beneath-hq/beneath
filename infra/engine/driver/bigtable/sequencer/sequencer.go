package sequencer

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/bigtable"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrTimeout is returned by CommitBatch and indicatets that the batch was expired
	ErrTimeout = errors.New("TimeoutError")
)

const (
	// Default options
	defaultTTL           = 60 * time.Second
	defaultMaxDrift      = 5 * time.Minute
	defaultMaxTTLDrift   = 10 * time.Second
	defaultCompactEveryN = 100

	// Next sequence number columns
	nextColumnFamily = "n"
	nextNumberColumn = "n"
	nextBatchColumn  = "c"

	// Columns describing stable state
	stableColumnFamily = "s"
	stableNumberColumn = "n"
	stableCountColumn  = "c"

	// Family storing commits of stable
	commitsColumnFamily = "c"

	// Serialization sizes
	serializedTimeSize  = 8
	serializedRangeSize = 16
)

// Sequencer generates global sequences with commit-tracking using BigTable
type Sequencer struct {
	// Table used to store sequences
	Table *bigtable.Table

	// TTL is the maximum amount of time that may pass between calls to BeginBatch and CommitBatch
	TTL time.Duration

	// MaxDrift is the bound of time we'll factor in for possible clock drift on the local machine
	MaxDrift time.Duration

	// MaxTTLDrift is the bound of time we'll factor in for possible clock drift locally over the course of TTL
	MaxTTLDrift time.Duration

	// CompactEveryN is the number of batches to wait before running compaction on a sequence
	CompactEveryN int64
}

// New initializes a Sequencer
func New(admin *bigtable.AdminClient, client *bigtable.Client, tableName string) (Sequencer, error) {
	createTable(admin, tableName)
	createColumnFamily(admin, tableName, nextColumnFamily, bigtable.MaxVersionsPolicy(1))
	createColumnFamily(admin, tableName, stableColumnFamily, bigtable.MaxVersionsPolicy(1))
	createColumnFamily(admin, tableName, commitsColumnFamily, bigtable.MaxVersionsPolicy(1))
	return Sequencer{
		Table:         client.Open(tableName),
		TTL:           defaultTTL,
		MaxDrift:      defaultMaxDrift,
		MaxTTLDrift:   defaultMaxTTLDrift,
		CompactEveryN: defaultCompactEveryN,
	}, nil
}

// create a table in BigTable if it doesn't already exist
func createTable(admin *bigtable.AdminClient, name string) {
	// create table
	err := admin.CreateTable(context.Background(), name)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating table '%s': %v", name, err))
		}
	}
}

// create column family in BigTable if it doesn't already exist
func createColumnFamily(admin *bigtable.AdminClient, tableName, cfName string, policy bigtable.GCPolicy) {
	// create column family
	err := admin.CreateColumnFamily(context.Background(), tableName, cfName)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating column family '%s': %v", cfName, err))
		}
	} else {
		// no error, so first time it's created -- set gc policy
		err = admin.SetGCPolicy(context.Background(), tableName, cfName, policy)
		if err != nil {
			panic(fmt.Errorf("error setting gc policy: %v", err))
		}
	}
}

// BeginBatch allocates a new batch of sequence numbers. You must call CommitBatch when it's spent.
func (s Sequencer) BeginBatch(ctx context.Context, key string, n int) (Batch, error) {
	// begin ttl
	now := time.Now()

	// create increment
	rmw := bigtable.NewReadModifyWrite()
	rmw.Increment(nextColumnFamily, nextBatchColumn, 1)
	rmw.Increment(nextColumnFamily, nextNumberColumn, int64(n))

	// apply increment
	row, err := s.Table.ApplyReadModifyWrite(ctx, key, rmw)
	if err != nil {
		return Batch{}, err
	}

	// read increment values
	var batch int64
	var number int64
	var btTime time.Time
	for _, col := range row[nextColumnFamily] {
		switch stripColumnFamily(col.Column, nextColumnFamily) {
		case nextBatchColumn:
			batch = bytesToInt(col.Value)
		case nextNumberColumn:
			number = bytesToInt(col.Value)
			btTime = col.Timestamp.Time()
		}
	}

	// return new batch
	return Batch{
		BatchID: batch - 1,
		Key:     key,
		Range: Range{
			From: number - int64(n),
			To:   number,
		},
		BigtableTime: btTime,
		LocalTime:    now,
	}, nil
}

// CommitBatch commits the batch, which facilitates accurate reporting in GetStable and GetCount.
// It will error with TimeoutError if you call it longer than TTL after BeginBatch.
func (s Sequencer) CommitBatch(ctx context.Context, batch Batch) error {
	// check timeout against local clock to catch early
	if time.Since(batch.LocalTime) >= (s.TTL - s.MaxTTLDrift) {
		return ErrTimeout
	}

	// create commit (we're using readmodifywrite to get server timestamp back)
	ckey := string(batch.Range.Bytes())
	rmw := bigtable.NewReadModifyWrite()
	// saving batch time so that compaction can detect if the commit was timed out
	rmw.AppendValue(commitsColumnFamily, ckey, timeToBytes(batch.BigtableTime))

	// apply rmw
	row, err := s.Table.ApplyReadModifyWrite(ctx, batch.Key, rmw)
	if err != nil {
		return err
	}

	// definitive timeout check: return timeout error if not within ttl
	col := row[commitsColumnFamily][0]
	if col.Timestamp.Time().Sub(batch.BigtableTime) >= s.TTL {
		// Append too slow -- it won't be counted during compaction
		return ErrTimeout
	}

	// run compaction every N batches (+1 to not run on 0th batch)
	if (batch.BatchID+1)%s.CompactEveryN == 0 {
		_, err := s.compactKey(ctx, batch.Key, false)
		if err != nil {
			// maybe just log and ignore? -- after all, batch was committed successfully
			return err
		}
	}

	return nil
}

// ClearState deletes all state related to a sequence
func (s Sequencer) ClearState(ctx context.Context, key string) error {
	mut := bigtable.NewMutation()
	mut.DeleteRow()
	return s.Table.Apply(ctx, key, mut)
}

// GetState returns the current state of the sequence
func (s Sequencer) GetState(ctx context.Context, key string) (State, error) {
	return s.compactKey(ctx, key, true)
}

// compactKey reads and compacts the sequence for key.
// If dry == true, the compacted state is reported, but not persisted.
func (s Sequencer) compactKey(ctx context.Context, key string, dry bool) (State, error) {
	// TODO: Prevent duplicate work on very high-throughput with a lock by key (and skip if locked)?

	// prep res
	state := State{}

	// read row
	row, err := s.Table.ReadRow(ctx, key, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return state, err
	}

	// if row doesn't exist, return empty state
	if len(row) == 0 {
		return state, nil
	}

	// set next
	for _, col := range row[nextColumnFamily] {
		switch stripColumnFamily(col.Column, nextColumnFamily) {
		case nextNumberColumn:
			state.Next = bytesToInt(col.Value)
			state.LatestTime = col.Timestamp.Time()
		case nextBatchColumn:
			state.NextBatch = bytesToInt(col.Value)
		}
	}

	// set current stable
	var oldStableNumberValue []byte
	for _, col := range row[stableColumnFamily] {
		switch stripColumnFamily(col.Column, stableColumnFamily) {
		case stableNumberColumn:
			oldStableNumberValue = col.Value
			state.NextStable = bytesToInt(col.Value) + 1
			state.StableTime = col.Timestamp.Time()
		case stableCountColumn:
			state.StableCount = bytesToInt(col.Value)
		}
	}

	// parse commits
	commits := parseCommits(row[commitsColumnFamily])
	state.CommitsPendingCompaction = int64(len(commits))

	// extract safe now time, i.e. greater of (local now - max drift) and (highest commit time)
	now := time.Now().Add(-1 * s.MaxDrift)
	for _, com := range commits {
		if com.CommitTime.After(now) {
			now = com.CommitTime
		}
	}

	// advance stable
	for _, com := range commits {
		// skip commits that are lower than stable
		// they are double-commits or timeouted commits
		if com.Range.From < state.NextStable {
			continue
		}

		// if commit is higher than stable, there's a gap (missing commit)
		if com.Range.From > state.NextStable {
			// break if the commit's batch was begun within the TTL because then we are now in unstable territory
			if now.Sub(com.BatchTime) < s.TTL {
				break
			}
			// The missing commit is timed out, this commit is also stable.
			// Advance stable to the start of this commit, so processing continues.
			state.NextStable = com.Range.From
		}

		// if the commit naturally extends the stable state, incorporate it
		if com.Range.From == state.NextStable {
			// advance Stable and StableTime
			state.NextStable = com.Range.To
			if com.CommitTime.After(state.StableTime) {
				state.StableTime = com.CommitTime
			}

			// if commit was not timed out, also include the count in committed
			if com.CommitTime.Sub(com.BatchTime) < s.TTL {
				state.StableCount += com.Range.To - com.Range.From
			}
		}
	}

	// save changes
	if !dry {
		// delete all commits that are now stable
		mut := bigtable.NewMutation()
		for _, com := range commits {
			if com.Range.To <= state.NextStable {
				mut.DeleteCellsInColumn(commitsColumnFamily, com.ColumnKey)
				state.CommitsCompacted++
				state.CommitsPendingCompaction--
			} else {
				break
			}
		}

		// update stable number and stable count
		mut.Set(stableColumnFamily, stableNumberColumn, bigtable.Time(state.StableTime), intToBytes(state.NextStable-1))
		mut.Set(stableColumnFamily, stableCountColumn, bigtable.Time(state.StableTime), intToBytes(state.StableCount))

		// only continue if anything actually changes
		if state.CommitsCompacted > 0 {
			// only write if another compaction hasn't already occurred
			// (note: this cond mutation is the best I could come up with that handles the edge-case of two simultaneous compactions
			// where there's two commits that happen to have the same timestamp)
			if len(oldStableNumberValue) == 0 {
				// no previous value, write if nothing has been written in the mean time
				cond := bigtable.FamilyFilter(stableColumnFamily)
				mut = bigtable.NewCondMutation(cond, nil, mut)
			} else {
				// we read a previous value, write if stable number is still the same
				cond := bigtable.ChainFilters(
					bigtable.FamilyFilter(stableColumnFamily),
					bigtable.ColumnFilter(stableNumberColumn),
					bigtable.LatestNFilter(1),
					// can't use bigtable.ValueFilter because it regexes
					bigtable.ValueRangeFilter(oldStableNumberValue, append(oldStableNumberValue, 0x00)),
				)
				mut = bigtable.NewCondMutation(cond, mut, nil)
			}

			// apply mut
			err := s.Table.Apply(ctx, key, mut)
			if err != nil {
				return State{}, err
			}
		}
	}

	return state, nil
}

// State represents the state of a sequence.
type State struct {
	// Next is the next number that will be issued with BeginBatch
	Next int64

	// NextBatch marks the next batch that will be issued with BeginBatch
	NextBatch int64

	// NextStable marks the lowest number that is not considered stable (all lower numbers have expired or been committed)
	NextStable int64

	// LatestTime marks the timestamp of the latest issued number
	LatestTime time.Time

	// StableTime marks the timestamp of the highest committed number
	StableTime time.Time

	// StableCount is the number of non-timeouted values below NextStable
	StableCount int64

	// CommitsPendingCompaction counts the number of commits not yet compacted
	CommitsPendingCompaction int64

	// CommitsCompacted counts the number of commits compacted in this call
	CommitsCompacted int64
}

// Range represents a sequence range
type Range struct {
	From int64
	To   int64
}

// NewRangeFromBytes recreates a Range from the output of Bytes()
func NewRangeFromBytes(b []byte) Range {
	if len(b) != serializedRangeSize {
		panic(fmt.Errorf("Not an encoded Range"))
	}
	return Range{
		From: bytesToInt(b[:8]),
		To:   bytesToInt(b[8:]),
	}
}

// Bytes encodes the range in 16 bytes
func (r Range) Bytes() []byte {
	res := append(intToBytes(r.From), intToBytes(r.To)...)
	// assert: len(res) == serializedRangeSize
	return res
}

// Batch represents a new batch of numbers
type Batch struct {
	BatchID      int64
	Key          string
	Range        Range
	BigtableTime time.Time
	LocalTime    time.Time
}

// NumberAtOffset returns the sequence numer at offset in the batch
func (b Batch) NumberAtOffset(offset int) int64 {
	res := b.Range.From + int64(offset)
	if res < b.Range.To {
		return res
	}
	panic(fmt.Errorf("OffsetAtIndex called with idx out of bounds"))
}

// Commit represents a committed batch
type Commit struct {
	ColumnKey  string
	Range      Range
	BatchTime  time.Time
	CommitTime time.Time
}

// reads items from commitsColumnFamily into Commit structs
func parseCommits(items []bigtable.ReadItem) []Commit {
	// build commits
	commits := make([]Commit, len(items))
	for idx, item := range items {
		// Value is serializedTimeSize. Due to use of append, multiple commits could cause extra bytes (hence the slicing).
		ckey := stripColumnFamily(item.Column, commitsColumnFamily)
		val := item.Value[:serializedTimeSize]
		com := Commit{
			ColumnKey:  ckey,
			Range:      NewRangeFromBytes([]byte(ckey)),
			BatchTime:  bytesToTime(val),
			CommitTime: item.Timestamp.Time(),
		}
		commits[idx] = com
	}

	// sort by range start
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Range.From < commits[j].Range.From
	})

	return commits
}

// encodes an int for storage in bigtable
func intToBytes(x int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, x)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// decodes an int encoded with intToBytes
func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func timeToBytes(t time.Time) []byte {
	return intToBytes(int64(t.UnixNano() / 1e3))
}

func bytesToTime(bs []byte) time.Time {
	return time.Unix(0, bytesToInt([]byte(bs))*1e3)
}

func stripColumnFamily(ckey string, cf string) string {
	// column keys returned by read are "cf:ckey", so we strip the "cf:"
	return ckey[len(cf)+1:]
}

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

	// EmptyValue is used to store just keys (without values) in Bigtable
	EmptyValue = []byte{0x00}
)

const (
	// TTL is the maximum amount of time that may pass between calls to BeginBatch and CommitBatch
	TTL = 60 * time.Second

	// MaxDrift is the bound of time we'll factor in for possible clock drift on the local machine
	MaxDrift = 5 * time.Minute

	// MaxTTLDrift is the bound of time we'll factor in for possible clock drift locally over the course of TTL
	MaxTTLDrift = 10 * time.Second

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

	// Run compaction every N batches
	compactEveryN = 100
)

// Sequencer generates global sequences with commit-tracking using BigTable
type Sequencer struct {
	Table *bigtable.Table
}

// New initializes a Sequencer
func New(admin *bigtable.AdminClient, client *bigtable.Client, tableName string) (Sequencer, error) {
	createTable(admin, tableName)
	createColumnFamily(admin, tableName, nextColumnFamily, bigtable.MaxVersionsPolicy(1))
	createColumnFamily(admin, tableName, stableColumnFamily, bigtable.MaxVersionsPolicy(1))
	createColumnFamily(admin, tableName, commitsColumnFamily, bigtable.MaxVersionsPolicy(1))
	return Sequencer{
		Table: client.Open(tableName),
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
		switch col.Column {
		case nextBatchColumn:
			batch = bytesToInt(col.Value)
		case nextNumberColumn:
			number = bytesToInt(col.Value)
			btTime = col.Timestamp.Time()
		}
	}

	// return new batch
	return Batch{
		BatchID: batch,
		Key:     key,
		Range: Range{
			// adding 1 to avoid 0 being used as an offset
			From: number - int64(n) + 1,
			To:   number + 1,
		},
		BigtableTime: btTime,
		LocalTime:    now,
	}, nil
}

// CommitBatch commits the batch, which facilitates accurate reporting in GetStable and GetCount.
// It will error with TimeoutError if you call it longer than TTL after BeginBatch.
func (s Sequencer) CommitBatch(ctx context.Context, batch Batch) error {
	// check timeout against local clock to catch early
	if time.Since(batch.LocalTime) >= (TTL - MaxTTLDrift) {
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
	if col.Timestamp.Time().Sub(batch.BigtableTime) >= TTL {
		// Append too slow -- it won't be counted during compaction
		return ErrTimeout
	}

	// run compaction every N batches
	if batch.BatchID%compactEveryN == 0 {
		_, err := s.compactKey(ctx, batch.Key, false)
		if err != nil {
			// maybe just log and ignore? -- after all, batch was committed successfully
			return err
		}
	}

	return nil
}

// GetState returns the current state of the sequence
func (s Sequencer) GetState(ctx context.Context, key string) (State, error) {
	return s.compactKey(ctx, key, true)
}

// compactKey reads and compacts the sequence for key.
// If dry == true, the compacted state is reported, but not persisted.
func (s Sequencer) compactKey(ctx context.Context, key string, dry bool) (State, error) {
	// prep res
	state := State{
		// Latest int64
		// LatestTime time.Time
		// Batches int64
		// Stable int64
		// StableTime time.Time
		// Committed int64
	}

	// read row
	row, err := s.Table.ReadRow(ctx, key, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return state, err
	}

	// set latest
	for _, col := range row[nextColumnFamily] {
		switch col.Column {
		case nextNumberColumn:
			state.Latest = bytesToInt(col.Value)
			state.LatestTime = col.Timestamp.Time()
		case nextBatchColumn:
			state.Batches = bytesToInt(col.Value)
		}
	}

	// set current stable
	for _, col := range row[stableColumnFamily] {
		switch col.Column {
		case stableNumberColumn:
			state.Stable = bytesToInt(col.Value)
			state.StableTime = col.Timestamp.Time()
		case stableCountColumn:
			state.Committed = bytesToInt(col.Value)
		}
	}

	// parse commits
	commits := parseCommits(row[commitsColumnFamily])

	// extract safe now time
	// i.e. greater of (local now - max drift) and (highest commit time)
	now := time.Now().Add(-1 * MaxDrift)
	for _, com := range commits {
		if com.CommitTime.After(now) {
			now = com.CommitTime
		}
	}

	for _, com := range commits {
		// skip commits that are lower than stable
		// they are double-commits or timeouted commits
		if com.Range.From < state.Stable {
			continue
		}

		// if commit is higher than stable, there's a gap (missing commit)
		if com.Range.From > state.Stable {
			// break if the commit's batch was begun within the TTL because then we are now in unstable territory
			if now.Sub(com.BatchTime) < TTL {
				break
			}
			// The missing commit is timed out, this commit is also stable.
			// Advance stable to the start of this commit, so processing continues.
			state.Stable = com.Range.From
		}

		// if the commit naturally extends the stable state, incorporate it
		if com.Range.From == state.Stable {
			// advance Stable and StableTime
			state.Stable = com.Range.To
			if com.CommitTime.After(state.StableTime) {
				state.StableTime = com.CommitTime
			}

			// if commit was not timed out, also include the count in committed
			if com.CommitTime.Sub(com.BatchTime) < TTL {
				state.Committed += com.Range.To - com.Range.From
			}
		}

	}

	return state, nil
}

// State represents the state of a sequence.
type State struct {
	// Latest is the latest number issued with BeginBatch
	Latest int64

	// LatestTime marks the timestamp of the latest issued number
	LatestTime time.Time

	// Batches counts the number of batches begun
	Batches int64

	// Stable marks the highest stable number (all lower numbers have expired or been committed)
	Stable int64

	// StableTime marks the timestamp associated with the highest stable number
	StableTime time.Time

	// Committed counts the number of committed values
	Committed int64
}

// Range represents a sequence range
type Range struct {
	From int64
	To   int64
}

// NewRangeFromBytes recreates a Range from the output of Bytes()
func NewRangeFromBytes(b []byte) Range {
	if len(b) != serializedRangeSize {
		panic(fmt.Errorf("Not an encoded OffsetSegment"))
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
		val := item.Value[:serializedTimeSize]
		com := Commit{
			Range:      NewRangeFromBytes(item.Value),
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

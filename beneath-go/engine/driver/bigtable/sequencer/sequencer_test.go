package sequencer

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func init() {
	os.Setenv("BIGTABLE_PROJECT_ID", "")
	os.Setenv("BIGTABLE_EMULATOR_HOST", "localhost:8086")
}

func setupTest(t *testing.T) (Sequencer, func(t *testing.T)) {
	// prepare BigTable client
	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, "", "")
	assert.NoError(t, err)
	admin, err := bigtable.NewAdminClient(ctx, "", "")
	assert.NoError(t, err)

	// create sequencer
	seq, err := New(admin, client, "sequencer_test")
	assert.NoError(t, err)

	// teardown func
	teardown := func(t *testing.T) {
		err := admin.DeleteTable(ctx, "sequencer_test")
		assert.NoError(t, err)
	}

	// done
	return seq, teardown
}

func TestCompaction(t *testing.T) {
	ctx := context.Background()
	seq, teardown := setupTest(t)
	defer teardown(t)

	// reconfigure sequencer for easier testing
	seq.CompactEveryN = 3

	// test naive
	k1 := "k1"
	b1, err := seq.BeginBatch(ctx, k1, 50)
	assert.NoError(t, err)
	s1, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	err = seq.CommitBatch(ctx, b1)
	assert.NoError(t, err)
	s2, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), b1.Range.From)
	assert.Equal(t, int64(50), b1.Range.To)
	assert.Equal(t, int64(3), b1.NumberAtOffset(3))
	assert.Panics(t, func() { b1.NumberAtOffset(50) })
	assert.Equal(t, int64(50), s1.Next)
	assert.Equal(t, int64(1), s1.NextBatch)
	assert.Equal(t, int64(0), s1.NextStable)
	assert.Equal(t, int64(0), s1.StableCount)
	assert.Equal(t, int64(0), s1.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s1.CommitsCompacted)
	assert.Equal(t, int64(50), s2.Next)
	assert.Equal(t, int64(1), s2.NextBatch)
	assert.Equal(t, int64(50), s2.NextStable)
	assert.Equal(t, int64(50), s2.StableCount)
	assert.Equal(t, int64(1), s2.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s2.CommitsCompacted)

	// test normal compaction
	b2, err := seq.BeginBatch(ctx, k1, 50)
	err = seq.CommitBatch(ctx, b2)
	assert.NoError(t, err)
	s3, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	b3, err := seq.BeginBatch(ctx, k1, 50)
	err = seq.CommitBatch(ctx, b3)
	assert.NoError(t, err)
	s4, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), b2.Range.From)
	assert.Equal(t, int64(100), b2.Range.To)
	assert.Equal(t, int64(100), s3.Next)
	assert.Equal(t, int64(100), s3.NextStable)
	assert.Equal(t, int64(2), s3.CommitsPendingCompaction)
	assert.Equal(t, int64(100), b3.Range.From)
	assert.Equal(t, int64(150), b3.Range.To)
	assert.Equal(t, int64(150), s4.Next)
	assert.Equal(t, int64(150), s4.NextStable)
	assert.Equal(t, int64(0), s4.CommitsPendingCompaction)
}

func TestCompactionOutOfOrder(t *testing.T) {
	ctx := context.Background()
	seq, teardown := setupTest(t)
	defer teardown(t)

	// reconfigure sequencer for easier testing
	seq.CompactEveryN = 3

	k1 := "k1"
	b1, err := seq.BeginBatch(ctx, k1, 5)
	assert.NoError(t, err)
	b2, err := seq.BeginBatch(ctx, k1, 10)
	assert.NoError(t, err)
	b3, err := seq.BeginBatch(ctx, k1, 15)
	assert.NoError(t, err)

	// expected: commit 1, commit 3, compaction, commit 2, manual compaction
	err = seq.CommitBatch(ctx, b1)
	assert.NoError(t, err)
	s1, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	err = seq.CommitBatch(ctx, b3)
	assert.NoError(t, err)
	s2, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	err = seq.CommitBatch(ctx, b2)
	assert.NoError(t, err)
	s3, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)
	s4, err := seq.compactKey(ctx, k1, false)
	assert.NoError(t, err)

	assert.Equal(t, int64(30), s1.Next)
	assert.Equal(t, int64(5), s1.NextStable)
	assert.Equal(t, int64(5), s1.StableCount)
	assert.Equal(t, int64(1), s1.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s1.CommitsCompacted)
	assert.Equal(t, int64(30), s2.Next)
	assert.Equal(t, int64(5), s2.NextStable)
	assert.Equal(t, int64(5), s2.StableCount)
	assert.Equal(t, int64(1), s2.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s2.CommitsCompacted)
	assert.Equal(t, int64(30), s3.Next)
	assert.Equal(t, int64(30), s3.NextStable)
	assert.Equal(t, int64(30), s3.StableCount)
	assert.Equal(t, int64(2), s3.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s3.CommitsCompacted)
	assert.Equal(t, int64(30), s4.Next)
	assert.Equal(t, int64(30), s4.NextStable)
	assert.Equal(t, int64(30), s4.StableCount)
	assert.Equal(t, int64(0), s4.CommitsPendingCompaction)
	assert.Equal(t, int64(2), s4.CommitsCompacted)
}

func TestTimeoutCompaction(t *testing.T) {
	ctx := context.Background()
	seq, teardown := setupTest(t)
	defer teardown(t)

	// reconfigure sequencer for easier testing
	seq.MaxDrift = time.Millisecond
	seq.MaxTTLDrift = -500 * time.Millisecond // -500ms takes the local test (an optimization) at the top of CommitBatch out of the equation
	seq.TTL = 100 * time.Millisecond
	seq.CompactEveryN = 3

	k1 := "k1"
	b1, err := seq.BeginBatch(ctx, k1, 5)
	assert.NoError(t, err)
	b2, err := seq.BeginBatch(ctx, k1, 10)
	assert.NoError(t, err)
	b3, err := seq.BeginBatch(ctx, k1, 15)
	assert.NoError(t, err)

	// approach: commit 1, commit 3, compaction, wait ttl, commit 2, timeout error, manual compaction, test tidy

	err = seq.CommitBatch(ctx, b1)
	assert.NoError(t, err)
	err = seq.CommitBatch(ctx, b3)
	assert.NoError(t, err)

	time.Sleep(seq.TTL)
	err = seq.CommitBatch(ctx, b2)
	assert.Equal(t, ErrTimeout, err)

	s1, err := seq.GetState(ctx, k1)
	assert.NoError(t, err)

	s2, err := seq.compactKey(ctx, k1, false)
	assert.NoError(t, err)

	assert.Equal(t, int64(30), s1.Next)
	assert.Equal(t, int64(30), s1.NextStable)
	assert.Equal(t, int64(20), s1.StableCount)
	assert.Equal(t, int64(2), s1.CommitsPendingCompaction)
	assert.Equal(t, int64(0), s1.CommitsCompacted)
	assert.Equal(t, int64(30), s2.Next)
	assert.Equal(t, int64(30), s2.NextStable)
	assert.Equal(t, int64(20), s2.StableCount)
	assert.Equal(t, int64(0), s2.CommitsPendingCompaction)
	assert.Equal(t, int64(2), s2.CommitsCompacted)
}

func TestStress(t *testing.T) {
	ctx := context.Background()
	seq, teardown := setupTest(t)
	defer teardown(t)

	seq.TTL = 1000 * time.Millisecond
	seq.MaxDrift = time.Millisecond
	seq.MaxTTLDrift = time.Millisecond
	seq.CompactEveryN = 5

	n := 1000
	key := "aaa"
	batchSize := 1000
	timeoutEveryN := 10

	// simulate many simultaneous ops
	group := new(errgroup.Group)
	for i := 0; i < n; i++ {
		i := i
		group.Go(func() error {
			batch, err := seq.BeginBatch(ctx, key, batchSize)
			if err != nil {
				return err
			}

			// maybe timeout
			if i%timeoutEveryN == 0 {
				return nil
			}

			return seq.CommitBatch(ctx, batch)
		})
	}

	// wait for all to finish
	err := group.Wait()
	assert.NoError(t, err)

	// expectations
	expectedNext := int64(n * batchSize)
	expectedBatches := int64(n)
	expectedStable := int64(expectedNext)
	expectedStableCount := int64(expectedNext - int64((n*batchSize)/timeoutEveryN))

	// read before timeouted values timeout
	// at this point, expecting to see lower values than expected after ttl
	s1, err := seq.GetState(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, expectedNext, s1.Next)
	assert.Equal(t, expectedBatches, s1.NextBatch)
	assert.Greater(t, expectedStable, s1.NextStable)
	assert.Greater(t, expectedStableCount, s1.StableCount)
	assert.NotEqual(t, int64(0), s1.CommitsPendingCompaction)

	// ensure everything times out
	time.Sleep(seq.TTL + seq.MaxDrift)

	// get state
	s2, err := seq.GetState(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, expectedNext, s2.Next)
	assert.Equal(t, expectedBatches, s2.NextBatch)
	assert.Equal(t, expectedStable, s2.NextStable)
	assert.Equal(t, expectedStableCount, s2.StableCount)
	assert.NotEqual(t, int64(0), s2.CommitsPendingCompaction)

	// trigger compaction
	s3, err := seq.compactKey(ctx, key, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedNext, s3.Next)
	assert.Equal(t, expectedBatches, s3.NextBatch)
	assert.Equal(t, expectedStable, s3.NextStable)
	assert.Equal(t, expectedStableCount, s3.StableCount)
	assert.Equal(t, int64(0), s3.CommitsPendingCompaction)
}

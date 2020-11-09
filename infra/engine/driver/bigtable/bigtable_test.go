package bigtable

import "testing"

func TestBigtable(t *testing.T) {
	// TODO: Write tests
	// Compacted reads
	// Non-compacted reads
	// Filtered reads
	// Secondary index reads
	// Writing where secondary indexes update works (doesn't create garbage on sequential writes)
	// Cleans up garbage secondary index data on reads

	// read uncompacted scenarios:
	// - 1st cursor covers missing segment, 2nd cursor covers present segment
	// - 1st cursor covers partially missing segment, 2nd cursor covers a totally missing segment, no third cursor
	// - an open-ended cursor should stay open, both if it finds data and if it doesn't
	// - 1st and 2nd cursors cover missing segments, third cursor covers present segment (invariant violation?)
}

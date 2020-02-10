package bigtable

import (
	"context"
	"encoding/binary"

	"cloud.google.com/go/bigtable"

	pb "github.com/beneath-core/beneath-go/engine/proto"
)

// CommitUsage implements engine.LookupService
func (b BigTable) CommitUsage(ctx context.Context, key []byte, usage pb.QuotaUsage) error {
	rmw := bigtable.NewReadModifyWrite()

	if usage.ReadOps > 0 {
		rmw.Increment(usageColumnFamilyName, usageReadOpsColumnName, usage.ReadOps)
		rmw.Increment(usageColumnFamilyName, usageReadRowsColumnName, usage.ReadRecords)
		rmw.Increment(usageColumnFamilyName, usageReadBytesColumnName, usage.ReadBytes)
	}

	if usage.WriteOps > 0 {
		rmw.Increment(usageColumnFamilyName, usageWriteOpsColumnName, usage.WriteOps)
		rmw.Increment(usageColumnFamilyName, usageWriteRowsColumnName, usage.WriteRecords)
		rmw.Increment(usageColumnFamilyName, usageWriteBytesColumnName, usage.WriteBytes)
	}

	_, err := b.Usage.ApplyReadModifyWrite(ctx, string(key), rmw)
	return err
}

// ReadSingleUsage implements engine.LookupService
func (b BigTable) ReadSingleUsage(ctx context.Context, key []byte) (pb.QuotaUsage, error) {
	// add filter for latest cell value
	filter := bigtable.RowFilter(bigtable.LatestNFilter(1))

	// read row
	row, err := b.Usage.ReadRow(ctx, string(key), filter)
	if err != nil {
		return pb.QuotaUsage{}, err
	}

	// build usage
	usage := b.rowToUsage(row)

	// done
	return usage, nil
}

// ReadUsage implements engine.LookupService
func (b BigTable) ReadUsage(ctx context.Context, fromKey []byte, toKey []byte, fn func(key []byte, usage pb.QuotaUsage) error) error {
	// convert keyRange to RowSet
	rr := bigtable.NewRange(string(fromKey), string(toKey))

	// define callback triggered on each bigtable row
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// get values
		key := []byte(row.Key())
		usage := b.rowToUsage(row)

		// trigger callback
		cbErr = fn(key, usage)
		if cbErr != nil {
			return false // stop
		}

		// continue
		return true
	}

	// read rows
	err := b.Usage.ReadRows(ctx, rr, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return err
	} else if cbErr != nil {
		return cbErr
	}

	// done
	return nil
}

func (b *BigTable) rowToUsage(row bigtable.Row) pb.QuotaUsage {
	usage := pb.QuotaUsage{}
	for _, cell := range row[usageColumnFamilyName] {
		colName := cell.Column[len(usageColumnFamilyName)+1:]
		switch colName {
		case usageReadOpsColumnName:
			usage.ReadOps = int64(binary.BigEndian.Uint64(cell.Value))
		case usageReadRowsColumnName:
			usage.ReadRecords = int64(binary.BigEndian.Uint64(cell.Value))
		case usageReadBytesColumnName:
			usage.ReadBytes = int64(binary.BigEndian.Uint64(cell.Value))
		case usageWriteOpsColumnName:
			usage.WriteOps = int64(binary.BigEndian.Uint64(cell.Value))
		case usageWriteRowsColumnName:
			usage.WriteRecords = int64(binary.BigEndian.Uint64(cell.Value))
		case usageWriteBytesColumnName:
			usage.WriteBytes = int64(binary.BigEndian.Uint64(cell.Value))
		}
	}
	return usage
}

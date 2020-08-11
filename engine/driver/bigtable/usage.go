package bigtable

import (
	"context"
	"encoding/binary"
	"time"

	"cloud.google.com/go/bigtable"
	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// CommitUsage implements engine.LookupService
func (b BigTable) CommitUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time, usage pb.QuotaUsage) error {
	// get table
	table := b.Usage
	if period > timeutil.PeriodDay {
		table = b.UsageTemp
	}

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

	if usage.ScanOps > 0 {
		rmw.Increment(usageColumnFamilyName, usageScanOpsColumnName, usage.ScanOps)
		rmw.Increment(usageColumnFamilyName, usageScanBytesColumnName, usage.ScanBytes)
	}

	key := makeUsageKey(id, period, ts, 0)
	_, err := table.ApplyReadModifyWrite(ctx, key, rmw)
	return err
}

// ClearUsage clears all usage data saved for the id
func (b BigTable) ClearUsage(ctx context.Context, id uuid.UUID) error {
	// clear usage table
	err := b.Admin.DropRowRange(ctx, usageTableName, string(id.Bytes()))
	if err != nil {
		return err
	}

	// clear temp usage table
	err = b.Admin.DropRowRange(ctx, usageTempTableName, string(id.Bytes()))
	if err != nil {
		return err
	}

	return nil
}

// ReadSingleUsage implements engine.LookupService
func (b BigTable) ReadSingleUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time) (pb.QuotaUsage, error) {
	// get table
	table := b.Usage
	if period > timeutil.PeriodDay {
		table = b.UsageTemp
	}

	// add filter for latest cell value
	filter := bigtable.RowFilter(bigtable.LatestNFilter(1))

	// read row
	key := makeUsageKey(id, period, ts, 0)
	row, err := table.ReadRow(ctx, key, filter)
	if err != nil {
		return pb.QuotaUsage{}, err
	}

	// build usage
	usage := b.rowToUsage(row)

	// done
	return usage, nil
}

// ReadUsage implements engine.LookupService
func (b BigTable) ReadUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, from time.Time, until time.Time, fn func(ts time.Time, usage pb.QuotaUsage) error) error {
	// get table
	table := b.Usage
	if period > timeutil.PeriodDay {
		table = b.UsageTemp
	}

	// convert keyRange to RowSet
	fromKey := makeUsageKey(id, period, from, 0)
	toKey := makeUsageKey(id, period, until, time.Second)
	rr := bigtable.NewRange(fromKey, toKey)

	// define callback triggered on each bigtable row
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// get values
		_, _, ts := parseUsageKey(row.Key())
		usage := b.rowToUsage(row)

		// trigger callback
		cbErr = fn(ts, usage)
		if cbErr != nil {
			return false // stop
		}

		// continue
		return true
	}

	// read rows
	err := table.ReadRows(ctx, rr, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)))
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
		case usageScanOpsColumnName:
			usage.ScanOps = int64(binary.BigEndian.Uint64(cell.Value))
		case usageScanBytesColumnName:
			usage.ScanBytes = int64(binary.BigEndian.Uint64(cell.Value))
		}
	}
	return usage
}

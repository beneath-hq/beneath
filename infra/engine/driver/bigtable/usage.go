package bigtable

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/proto"
)

// WriteUsage implements engine.LookupService
func (b BigTable) WriteUsage(ctx context.Context, id uuid.UUID, label driver.UsageLabel, ts time.Time, usage pb.QuotaUsage) error {
	table := b.tableForUsageLabel(label)

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

	key := b.makeUsageKey(id, label, ts)
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

// ReadUsageSingle implements engine.LookupService
func (b BigTable) ReadUsageSingle(ctx context.Context, id uuid.UUID, label driver.UsageLabel, ts time.Time) (pb.QuotaUsage, error) {
	key := b.makeUsageKey(id, label, ts)
	table := b.tableForUsageLabel(label)
	filter := bigtable.RowFilter(bigtable.LatestNFilter(1))
	row, err := table.ReadRow(ctx, key, filter)
	if err != nil {
		return pb.QuotaUsage{}, err
	}
	return b.rowToUsage(row), nil
}

// ReadUsageRange implements engine.LookupService
func (b BigTable) ReadUsageRange(ctx context.Context, id uuid.UUID, label driver.UsageLabel, from time.Time, to time.Time, limit int, fn func(ts time.Time, usage pb.QuotaUsage) error) error {
	// convert keyRange to RowSet
	fromKey := b.makeUsageKey(id, label, from)
	toKey := b.makeUsageKey(id, label, to)
	rr := bigtable.NewRange(fromKey, toKey)

	// define callback triggered on each bigtable row
	var cbErr error
	cb := func(row bigtable.Row) bool {
		// get values
		ts := b.parseUsageKeyTime(row.Key())
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
	table := b.tableForUsageLabel(label)
	err := table.ReadRows(ctx, rr, cb, bigtable.RowFilter(bigtable.LatestNFilter(1)), bigtable.LimitRows(int64(limit)))
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

func (b *BigTable) tableForUsageLabel(label driver.UsageLabel) *bigtable.Table {
	if label == driver.UsageLabelHourly {
		return b.UsageTemp
	}
	return b.Usage
}

func (b *BigTable) usageLabelToByte(label driver.UsageLabel) byte {
	switch label {
	case driver.UsageLabelMonthly:
		return 'M'
	case driver.UsageLabelHourly:
		return 'H'
	case driver.UsageLabelQuotaMonth:
		return 'q'
	default:
		panic(fmt.Errorf("missing case in usageLabelToByte"))
	}
}

func (b *BigTable) makeUsageKey(id uuid.UUID, label driver.UsageLabel, ts time.Time) string {
	cap := uuid.Size + 1 + int64ByteSize
	key := make([]byte, 0, cap)
	key = append(key, id[:]...)
	key = append(key, b.usageLabelToByte(label))
	key = append(key, timeToBytesMs(ts)...)
	return byteSliceToString(key)
}

func (b *BigTable) parseUsageKeyTime(key string) time.Time {
	bs := stringToByteSlice(key)
	if len(bs) != uuid.Size+1+int64ByteSize {
		panic(fmt.Errorf("not a usage key"))
	}

	ts := bytesToTimeMs(bs[uuid.Size+1:])
	return ts
}

package bigtable

import (
	"context"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

// MaxKeySize implements beneath.Service
func (b BigTable) MaxKeySize() int {
	// Bigtable limit is 4096 bytes
	// Longest keys are instanceID#secondaryKey#primaryKey
	return 1024 // 1 kb
}

// MaxRecordSize implements beneath.Service
func (b BigTable) MaxRecordSize() int {
	// Bigtable goes up to 10 MB, but performance suffers
	// We have to be careful to support user requests of 1000 rows at once
	return 8192 // 8 kb
}

// MaxRecordsInBatch implements beneath.Service
func (b BigTable) MaxRecordsInBatch() int {
	return 10000
}

// RegisterInstance implements beneath.Service
func (b BigTable) RegisterInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	return nil
}

// RemoveInstance implements beneath.Service
func (b BigTable) RemoveInstance(ctx context.Context, s driver.Table, i driver.TableInstance) error {
	codec := s.GetCodec()

	// get table names
	logTable, indexesTable := logTableName, indexesTableName
	if logExpires(s) {
		logTable = logExpiringTableName
	}
	if indexExpires(s) {
		indexesTable = indexesExpiringTableName
	}

	if s.GetUseIndex() {
		// secondary indexes
		for _, index := range codec.SecondaryIndexes {
			indexID := index.GetIndexID()
			err := b.Admin.DropRowRange(ctx, indexesTable, string(indexID[:]))
			if err != nil {
				return err
			}
		}

		// primary index
		indexID := codec.PrimaryIndex.GetIndexID()
		err := b.Admin.DropRowRange(ctx, indexesTable, string(indexID[:]))
		if err != nil {
			return err
		}
	}

	// log
	if s.GetUseLog() {
		instanceID := i.GetTableInstanceID()
		err := b.Admin.DropRowRange(ctx, logTable, string(instanceID[:]))
		if err != nil {
			return err
		}

		// sequencer
		err = b.Sequencer.ClearState(ctx, makeSequencerKey(instanceID))
		if err != nil {
			return err
		}
	}

	return nil
}

// Reset implements beneath.Service
func (b BigTable) Reset(ctx context.Context) error {
	tables := []string{
		logTableName,
		logExpiringTableName,
		indexesTableName,
		indexesExpiringTableName,
		sequencerTableName,
		usageTableName,
	}

	for _, table := range tables {
		err := b.Admin.DropRowRange(ctx, table, "")
		if err != nil {
			return err
		}
	}

	return nil
}

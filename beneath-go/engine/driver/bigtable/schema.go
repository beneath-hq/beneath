package bigtable

import (
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	uuid "github.com/satori/go.uuid"
)

// row key for sequencer entries
func makeSequencerKey(instanceID uuid.UUID) string {
	return string(instanceID[:])
}

// row key for a log row (full)
func makeLogKey(instanceID uuid.UUID, offset int64, primaryKey []byte) string {
	key := tuple.Tuple{offset, primaryKey}.Pack()
	res := append(instanceID[:], key...)
	return byteSliceToString(res)
}

// row key prefix for a log row (excludes primary key details)
func makeLogKeyPrefix(instanceID uuid.UUID, offset int64) string {
	key := tuple.Tuple{offset}.Pack()
	res := append(instanceID[:], key...)
	return byteSliceToString(res)
}

// parses row keys returned by makeLogKey
func parseLogKey(key string) (instanceID uuid.UUID, offset int64, primaryKey []byte) {
	bs := stringToByteSlice(key)

	if len(bs) < uuid.Size {
		panic(fmt.Errorf("not an offset key"))
	}

	tpl, err := tuple.Unpack(bs[uuid.Size:])
	if err != nil {
		panic(err)
	} else if len(tpl) != 2 {
		panic(fmt.Errorf("not an offset key"))
	}

	return uuid.FromBytesOrNil(bs[:uuid.Size]), tpl[0].(int64), tpl[1].([]byte)
}

// mutation to insert log row
func makeLogInsert(rowTime time.Time, processTime time.Time, avro []byte) *bigtable.Mutation {
	ts := bigtable.Time(rowTime)
	mut := bigtable.NewMutation()
	mut.Set(logColumnFamilyName, logAvroColumnName, ts, avro)
	mut.Set(logColumnFamilyName, logProcessTimeColumnName, ts, timeToBytesMs(processTime))
	return mut
}

// row key for primary index rows
func makePrimaryIndexKey(instanceID uuid.UUID, primaryKey []byte) string {
	res := append(instanceID[:], primaryKey...)
	return byteSliceToString(res)
}

// parses row keys returned by makePrimaryIndexKey
func parsePrimaryIndexKey(key string) (instanceID uuid.UUID, primaryKey []byte) {
	if len(key) < uuid.Size {
		panic(fmt.Errorf("not a secondary index key"))
	}
	bs := stringToByteSlice(key)
	return uuid.FromBytesOrNil(bs[:uuid.Size]), bs[uuid.Size:]
}

// mutation to insert primary index row
func makePrimaryIndexInsert(rowTime time.Time, normalize bool, offset int64, avro []byte) *bigtable.Mutation {
	ts := bigtable.Time(rowTime)
	mut := bigtable.NewMutation()
	if normalize {
		mut.Set(indexesColumnFamilyName, indexesOffsetColumnName, ts, intToBytes(offset))
	} else {
		mut.Set(indexesColumnFamilyName, indexesAvroColumnName, ts, avro)
	}
	return mut
}

// row key for a secondary index row (full)
func makeSecondaryIndexKey(instanceID uuid.UUID, secondaryKey []byte, primaryKey []byte) string {
	key := tuple.Tuple{secondaryKey, primaryKey}.Pack()
	res := append(instanceID[:], key...)
	return byteSliceToString(res)
}

// row key prefix for a secondary index row (excludes primary key details)
func makeSecondaryIndexKeyPrefix(instanceID uuid.UUID, secondaryKey []byte) string {
	key := tuple.Tuple{secondaryKey}.Pack()
	res := append(instanceID[:], key...)
	return byteSliceToString(res)
}

// parses row keys returned by makeSecondaryIndexKey
func parseSecondaryIndexKey(key string) (instanceID uuid.UUID, secondaryKey []byte, primaryKey []byte) {
	bs := stringToByteSlice(key)

	if len(bs) < uuid.Size {
		panic(fmt.Errorf("not a secondary index key"))
	}

	tpl, err := tuple.Unpack(bs[uuid.Size:])
	if err != nil {
		panic(err)
	} else if len(tpl) != 2 {
		panic(fmt.Errorf("not a secondary index key"))
	}

	return uuid.FromBytesOrNil(bs[:uuid.Size]), tpl[0].([]byte), tpl[1].([]byte)
}

// mutation to insert secondary index row
func makeSecondaryIndexInsert(rowTime time.Time, normalize bool, offset int64, avro []byte) *bigtable.Mutation {
	ts := bigtable.Time(rowTime)
	mut := bigtable.NewMutation()
	// We currently always normalize secondary indexes to ensure consistency (see lookup.go for details)
	mut.Set(indexesColumnFamilyName, indexesOffsetColumnName, ts, intToBytes(offset))
	return mut
}

// mutation to delete secondary index row
func makeSecondaryIndexDelete(until time.Time, normalize bool) *bigtable.Mutation {
	mut := bigtable.NewMutation()
	// We currently always normalize secondary indexes to ensure consistency (see lookup.go for details)
	mut.DeleteTimestampRange(indexesColumnFamilyName, indexesOffsetColumnName, bigtable.Timestamp(0), bigtable.Time(until))
	return mut
}

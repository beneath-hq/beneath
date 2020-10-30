package bigtable

import (
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/codec/ext/tuple"
)

// IndexHash represents an index ID salted with an instance ID
type IndexHash [uuid.Size]byte

// row key for sequencer entries
func makeSequencerKey(instanceID uuid.UUID) string {
	return string(instanceID[:])
}

// row key for a log row (full)
func makeLogKey(instanceID uuid.UUID, offset int64, primaryKey []byte) string {
	cap := uuid.Size + int64ByteSize + len(primaryKey)
	key := make([]byte, 0, cap)
	key = append(key, instanceID[:]...)
	key = append(key, intToBytes(offset)...)
	if primaryKey != nil {
		key = append(key, primaryKey...)
	}
	return byteSliceToString(key)
}

// parses row keys returned by makeLogKey
func parseLogKey(key string) (instanceID uuid.UUID, offset int64, primaryKey []byte) {
	bs := stringToByteSlice(key)
	if len(bs) < uuid.Size+int64ByteSize {
		panic(fmt.Errorf("not an offset key"))
	}

	instanceID = uuid.FromBytesOrNil(bs[:uuid.Size])
	offset = bytesToInt(bs[uuid.Size:(uuid.Size + int64ByteSize)])
	primaryKey = bs[(uuid.Size + int64ByteSize):]

	return instanceID, offset, primaryKey
}

func makeLogRowSetFromRange(instanceID uuid.UUID, from int64, to int64) bigtable.RowSet {
	if to > 0 {
		bk := makeLogKey(instanceID, from, nil)
		ek := makeLogKey(instanceID, to, nil)
		return bigtable.NewRange(bk, ek)
	}

	bk := makeLogKey(instanceID, from, nil)
	ek := byteSliceToString(tuple.PrefixSuccessor(instanceID[:]))
	return bigtable.NewRange(bk, ek)
}

// mutation to insert log row
func makeLogInsert(rowTime time.Time, processTime time.Time, avro []byte) *bigtable.Mutation {
	ts := bigtable.Time(rowTime)
	mut := bigtable.NewMutation()
	mut.Set(logColumnFamilyName, logAvroColumnName, ts, avro)
	mut.Set(logColumnFamilyName, logProcessTimeColumnName, ts, timeToBytesMs(processTime))
	return mut
}

// hashes a stream instance ID and a stream index ID to produce a unique, random prefix for indexed data
func makeIndexHash(instanceID uuid.UUID, indexID uuid.UUID) IndexHash {
	var res IndexHash
	for i := 0; i < uuid.Size; i++ {
		res[i] = instanceID[i] + indexID[i]
	}
	return res
}

// row key for primary and secondary index rows
func makeIndexKey(hash IndexHash, indexKey []byte) string {
	res := append(hash[:], indexKey...)
	return byteSliceToString(res)
}

// parses row keys created by makeIndexKey
func parseIndexKey(key string) (hash IndexHash, indexKey []byte) {
	if len(key) < uuid.Size {
		panic(fmt.Errorf("not an index key"))
	}
	bs := stringToByteSlice(key)
	copy(hash[:], bs[:uuid.Size])
	return hash, bs[uuid.Size:]
}

func makeIndexRowSetFromRange(hash IndexHash, kr codec.KeyRange) bigtable.RowSet {
	if kr.IsNil() {
		return bigtable.PrefixRange(makeIndexKey(hash, nil))
	} else if kr.RangeEnd == nil {
		bk := makeIndexKey(hash, kr.Base)
		ek := byteSliceToString(tuple.PrefixSuccessor(hash[:]))
		return bigtable.NewRange(bk, ek)
	} else {
		bk := makeIndexKey(hash, kr.Base)
		ek := makeIndexKey(hash, kr.RangeEnd)
		return bigtable.NewRange(bk, ek)
	}
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

// mutation to insert secondary index row
func makeSecondaryIndexInsert(rowTime time.Time, normalize bool, offset int64, avro []byte) *bigtable.Mutation {
	ts := bigtable.Time(rowTime)
	mut := bigtable.NewMutation()
	// We currently always normalize secondary indexes to ensure consistency (see lookup.go for details)
	// TODO: This doesn't fly if log retention is less than index retention
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

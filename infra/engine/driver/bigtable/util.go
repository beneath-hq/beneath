package bigtable

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	"github.com/beneath-hq/beneath/pkg/timeutil"
)

const (
	int64ByteSize = 8
)

// returns true if data in log for table expires; returns false if data should be persisted forever
func logExpires(s driver.Table) bool {
	return s.GetLogRetention() != time.Duration(0)
}

// returns true if data in indexes for table expires; returns false if data should be persisted forever
func indexExpires(s driver.Table) bool {
	return s.GetIndexRetention() != time.Duration(0)
}

// turns t into an expiration time by adding retention to t
func toPersistedTime(t time.Time, retention time.Duration) time.Time {
	if retention == 0 {
		return t
	}
	return t.Add(retention)
}

// reverses the effect of toPersistedTime
func fromPersistedTime(t time.Time, retention time.Duration) time.Time {
	if retention == 0 {
		return t
	}
	return t.Add(-1 * retention)
}

// unsafely casts []byte to a string (saving memory)
// do not mutate input after calling
func byteSliceToString(bs []byte) string {
	// TODO: optimize using unsafe
	return string(bs)
}

// unsafely casts string to []byte (saving memory)
// do not mutate output
func stringToByteSlice(s string) []byte {
	// TODO: optimize using unsafe
	return []byte(s)
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

// encodes a time for storage in bigtable
func timeToBytesMs(t time.Time) []byte {
	ms := timeutil.UnixMilli(t)
	return intToBytes(ms)
}

// decodes a time encoded with timeToBytesMs
func bytesToTimeMs(b []byte) time.Time {
	ms := bytesToInt(b)
	return timeutil.FromUnixMilli(ms)
}

// splitCommonPrefix returns the common prefix of a and b, as well as the respective remainders
func splitCommonPrefix(a []byte, b []byte) (prefix []byte, aa []byte, bb []byte) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	var i int
	for i = 0; i < n; i++ {
		if a[i] != b[i] {
			break
		}
	}

	return a[:i], a[i:], b[i:]
}

func stripColumnFamily(ckey string, cf string) string {
	// column keys returned by read are "cf:ckey", so we strip the "cf:"
	return ckey[len(cf)+1:]
}

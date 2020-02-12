package metrics

import (
	"time"

	"github.com/beneath-core/pkg/codec/ext/tuple"
	"github.com/beneath-core/pkg/timeutil"
	uuid "github.com/satori/go.uuid"
)

// metricsKey returns the binary representation of the key
func metricsKey(period timeutil.Period, entityID uuid.UUID, ts time.Time) []byte {
	// prefix
	prefix := append([]byte{period.Byte()}, entityID.Bytes()...)

	// timestamp
	ts = timeutil.Floor(ts, period)
	unix := ts.Unix()
	timeEncoded := tuple.Tuple{unix}.Pack()

	// append to metrics key prefix
	return append(prefix, timeEncoded...)
}

// decodeMetricsKey decodes the binary representation of the key
func decodeMetricsKey(key []byte) (period timeutil.Period, entityID uuid.UUID, ts time.Time) {
	period, err := timeutil.PeriodFromByte(key[0])
	if err != nil {
		panic(err)
	}

	entityID, err = uuid.FromBytes(key[1:17])
	if err != nil {
		panic(err)
	}

	tup, err := tuple.Unpack(key[17:])
	if err != nil {
		panic(err)
	}
	ts = time.Unix(tup[0].(int64), 0)

	return period, entityID, ts
}

package timeutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Floor truncates ts to the given period
func Floor(ts time.Time, p Period) time.Time {
	ts = ts.UTC()
	switch p {
	case PeriodMinute:
		return time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), 0, 0, time.UTC)
	case PeriodHour:
		return time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, time.UTC)
	case PeriodDay:
		return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
	case PeriodMonth:
		return time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, time.UTC)
	case PeriodYear:
		return time.Date(ts.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		panic(fmt.Errorf("unknown period '%d'", p))
	}
}

// UnixMilli converts t to milliseconds since 1970
func UnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// FromUnixMilli converts milliseconds since 1970 to a time.Time
func FromUnixMilli(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// ToBytes returns the time.Time in a binary representation
func ToBytes(t time.Time) []byte {
	b, err := t.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return b
}

// Parse makes a best effort to parse val as a time.Time.
// Input will typically come from a user.
func Parse(val interface{}, allowNil bool) (time.Time, error) {
	// check nil
	if val == nil {
		if allowNil {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("couldn't parse nil timestamp")
	}

	// try parsing as date format
	if str, ok := val.(string); ok {
		// try "2006-01-02T15:04:05Z07:00" (RFC3339)
		t, err := time.Parse(time.RFC3339, str)
		if err == nil {
			return t, nil
		}

		// try "2006-01-02T15:04:05"
		t, err = time.Parse("2006-01-02T15:04:05", str)
		if err == nil {
			return t, nil
		}

		// try "2006-01-02"
		t, err = time.Parse("2006-01-02", str)
		if err == nil {
			return t, nil
		}
	}

	// try parsing as milliseconds
	var ms int64
	var err error
	var errored bool
	switch msT := val.(type) {
	case int:
		ms = int64(msT)
	case int32:
		ms = int64(msT)
	case int64:
		ms = int64(msT)
	case float64:
		ms = int64(msT)
	case string:
		ms, err = strconv.ParseInt(msT, 0, 64)
		if err != nil {
			errored = true
		}
	case json.Number:
		ms, err = msT.Int64()
		if err != nil {
			errored = true
		}
	default:
		errored = true
	}

	// break if couldn't parse
	if errored {
		return time.Time{}, fmt.Errorf("couldn't parse '%v' as a timestamp", val)
	}

	// check in range for time.Unix
	if ms > 9223372036853 || ms < -9223372036853 { // 2^64/2/1000/1000
		return time.Time{}, fmt.Errorf("timestamp out of range (nanoseconds must fit in 64-bit integer)")
	}

	// done
	return time.Unix(0, ms*int64(time.Millisecond)), nil
}

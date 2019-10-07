package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"

	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
)

// GetUsage returns metrics packets for the given length of time
func GetUsage(ctx context.Context, entityID uuid.UUID, period string, from time.Time, until time.Time) ([]time.Time, []pb.QuotaUsage, error) {
	// if "until" is 0 (it wasn't provided), set it to the current time
	if until.IsZero() {
		if period == HourlyPeriod {
			until = HourlyTimeFloor(time.Now())
		} else if period == MonthlyPeriod {
			until = MonthlyTimeFloor(time.Now())
		} else {
			panic("the period needs to be either monthly or hourly")
		}
	}

	// check that the provided time range corresponds to keys in the table
	if period == HourlyPeriod {
		if (from != HourlyTimeFloor(from)) || (until != HourlyTimeFloor(until)) {
			return nil, nil, fmt.Errorf("the provided times must be the exact beginning of the hour")
		}
	} else if period == MonthlyPeriod {
		if (from != MonthlyTimeFloor(from)) || (until != MonthlyTimeFloor(until)) {
			return nil, nil, fmt.Errorf("the provided times must be the exact beginning of the month")
		}
	}

	// prevent the retrieval of too many packets at once
	numPeriods := countPeriods(period, from, until)
	if numPeriods > 100 {
		return nil, nil, fmt.Errorf("the given length of time is too long")
	}

	// create key range
	fromKey := metricsKey(period, entityID, from)
	toKey := metricsKey(period, entityID, until)
	toKey = tuple.Successor(toKey)

	// read usage table and collect usage metrics
	var usagePackets []pb.QuotaUsage
	var timePeriods []time.Time
	err := db.Engine.Tables.ReadUsageRange(ctx, fromKey, toKey, func(key []byte, u pb.QuotaUsage) error {
		usagePackets = append(usagePackets, u)
		_, _, timePeriod := decodeMetricsKey(key)
		timePeriods = append(timePeriods, timePeriod)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("Error reading from Metrics table: %s", err.Error()))
	}

	// get rid of empty usagePackets
	idx := 0
	for i, v := range usagePackets {
		if (v.ReadOps > 0) || (v.WriteOps > 0) {
			usagePackets[idx] = v
			timePeriods[idx] = timePeriods[i]
			idx++
		}
	}
	usagePackets = usagePackets[:idx]
	timePeriods = timePeriods[:idx]

	return timePeriods, usagePackets, nil
}

func countPeriods(period string, from time.Time, until time.Time) int {
	if period == HourlyPeriod {
		numPeriods := int(until.Sub(from).Hours())
		return numPeriods
	} else if period == MonthlyPeriod {
		untilYear, untilMonth, _ := until.Date()
		fromYear, fromMonth, _ := from.Date()
		numPeriods := (untilYear-fromYear)*12 + int(untilMonth) - int(fromMonth)
		return numPeriods
	} else {
		panic("the period needs to be either monthly or hourly")
	}
}

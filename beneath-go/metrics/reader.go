package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
)

// GetUsage returns metrics packets for the given length of time
func GetUsage(ctx context.Context, entityID uuid.UUID, period string, from time.Time, until time.Time) ([]pb.QuotaUsage, error) {
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
			return nil, fmt.Errorf("the provided time range must be exactly hourly")
		}
	} else if period == MonthlyPeriod {
		if (from != MonthlyTimeFloor(from)) || (until != MonthlyTimeFloor(until)) {
			return nil, fmt.Errorf("the provided time range must be exactly monthly")
		}
	}

	// prevent the retrieval of too many packets at once
	numPeriods := countPeriods(period, from, until)
	if numPeriods > 100 {
		return nil, fmt.Errorf("the given length of time is too long")
	}

	// create key range
	fromKey := metricsKey(period, entityID, from)
	toKey := metricsKey(period, entityID, until)

	// read usage table and collect usage metrics
	usagePackets := make([]pb.QuotaUsage, numPeriods)
	counter := 0
	err := db.Engine.Tables.ReadUsageRange(ctx, fromKey, toKey, func(key []byte, u pb.QuotaUsage) error {
		usagePackets[counter] = u
		counter++
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("Error reading from Metrics table: %s", err.Error()))
	}

	return usagePackets, nil
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

package metrics

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/engine/proto"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

const (
	maxPeriods = 175 // roughly one week on an hourly basis
)

// GetCurrentUsage returns an ID's usage for the current monthly period
func GetCurrentUsage(ctx context.Context, entityID uuid.UUID) pb.QuotaUsage {
	// load from bigtable
	usage, err := hub.Engine.ReadSingleUsage(ctx, entityID, timeutil.PeriodMonth, time.Now())
	if err != nil {
		panic(fmt.Errorf("error reading metrics: %s", err.Error()))
	}

	return usage
}

// GetHistoricalUsage returns usage info for the given length of time
func GetHistoricalUsage(ctx context.Context, entityID uuid.UUID, period timeutil.Period, from time.Time, until time.Time) ([]time.Time, []pb.QuotaUsage, error) {
	// check is supported period
	if period != timeutil.PeriodHour && period != timeutil.PeriodMonth {
		return nil, nil, fmt.Errorf("usage is calculated only in hourly and monthly periods")
	}

	// if "until" is 0, set it to the current time
	if until.IsZero() {
		until = timeutil.Floor(time.Now(), period)
	}

	// check that the provided time range corresponds to keys in the table
	if from != timeutil.Floor(from, period) || until != timeutil.Floor(until, period) {
		return nil, nil, fmt.Errorf("from and until must exactly match period beginnings")
	}

	// prevent the retrieval of too many rows at once
	if period.Count(from, until) > maxPeriods {
		return nil, nil, fmt.Errorf("time span too long")
	}

	// read usage table and collect usage metrics
	var times []time.Time
	var usages []pb.QuotaUsage
	err := hub.Engine.ReadUsage(ctx, entityID, period, from, until, func(ts time.Time, usage pb.QuotaUsage) error {
		times = append(times, ts)
		usages = append(usages, usage)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("error reading from metrics table: %s", err.Error()))
	}

	return times, usages, nil
}

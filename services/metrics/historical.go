package metrics

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/infrastructure/engine/driver"
	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
)

const (
	maxPeriods = 175 // roughly one week on an hourly basis
)

// GetHistoricalUsage returns usage info for the given length of time
func (s *Service) GetHistoricalUsage(ctx context.Context, entityID uuid.UUID, label driver.UsageLabel, from time.Time, until time.Time) ([]time.Time, []pb.QuotaUsage, error) {
	// if "until" is 0, set it to the current time
	if until.IsZero() {
		until = time.Now()
	}

	// read usage table and collect usage metrics
	var times []time.Time
	var usages []pb.QuotaUsage
	err := s.engine.Usage.ReadUsageRange(ctx, entityID, label, from, until, maxPeriods, func(ts time.Time, usage pb.QuotaUsage) error {
		times = append(times, ts)
		usages = append(usages, usage)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("error reading from metrics table: %s", err.Error()))
	}

	return times, usages, nil
}

package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/proto"
	uuid "github.com/satori/go.uuid"
)

const (
	maxPeriods = 175 // roughly one week on an hourly basis
)

// GetHistoricalUsageSingle returns usage info for a specific timestamp
func (s *Service) GetHistoricalUsageSingle(ctx context.Context, entityID uuid.UUID, label driver.UsageLabel, ts time.Time) (pb.QuotaUsage, error) {
	return s.engine.Usage.ReadUsageSingle(ctx, entityID, label, ts)
}

// GetHistoricalUsageRange returns usage info for the given length of time
func (s *Service) GetHistoricalUsageRange(ctx context.Context, entityID uuid.UUID, label driver.UsageLabel, from *time.Time, until *time.Time) ([]time.Time, []pb.QuotaUsage, error) {
	if from == nil {
		t := time.Time{}
		from = &t
	}
	if until == nil {
		t := time.Now()
		until = &t
	}

	// read and collect usage
	var times []time.Time
	var usages []pb.QuotaUsage
	err := s.engine.Usage.ReadUsageRange(ctx, entityID, label, *from, *until, maxPeriods, func(ts time.Time, usage pb.QuotaUsage) error {
		times = append(times, ts)
		usages = append(usages, usage)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("error reading usage: %s", err.Error()))
	}

	return times, usages, nil
}

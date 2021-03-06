package timeutil

import "time"

// FixedOffsetPeriod is a fixed-duration period with a custom epoch.
// Eg. [epoch=2020-11-01T01:02:03, duration=2 days] gives periods
// [2020-11-01T01:02:03, 2020-11-03T01:02:03, 2020-11-05T01:02:03, ...]
type FixedOffsetPeriod struct {
	Epoch          time.Time
	PeriodDuration time.Duration
}

// NewFixedOffsetPeriod creates a FixedOffsetPeriod
func NewFixedOffsetPeriod(epoch time.Time, periodDuration time.Duration) FixedOffsetPeriod {
	return FixedOffsetPeriod{
		Epoch:          epoch,
		PeriodDuration: periodDuration,
	}
}

// Floor returns the lowest time that is in the same period as ts
func (p FixedOffsetPeriod) Floor(ts time.Time) time.Time {
	return p.Epoch.Add(ts.Sub(p.Epoch).Truncate(p.PeriodDuration))
}

// PrevFloor returns the floor of the previous multiple of the given period of ts
func (p FixedOffsetPeriod) PrevFloor(ts time.Time) time.Time {
	return p.Floor(ts).Add(-1 * p.PeriodDuration)
}

// Next rounds ts up to the next multiple of the given period
func (p FixedOffsetPeriod) Next(ts time.Time) time.Time {
	return p.Floor(ts).Add(p.PeriodDuration)
}

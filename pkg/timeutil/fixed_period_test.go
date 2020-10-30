package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFloor(t *testing.T) {
	epoch := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	p1 := NewFixedOffsetPeriod(epoch, time.Hour)
	assert.Equal(t, time.Date(2020, 1, 2, 22, 4, 5, 0, time.UTC), p1.Floor(time.Date(2020, 1, 2, 23, 1, 1, 0, time.UTC)))

	p2 := NewFixedOffsetPeriod(epoch, 31*24*time.Hour)
	assert.Equal(t, epoch, p2.Floor(epoch.Add(10*24*time.Hour)))
	assert.Equal(t, epoch.Add(31*24*time.Hour), p2.Floor(epoch.Add(37*24*time.Hour)))
}

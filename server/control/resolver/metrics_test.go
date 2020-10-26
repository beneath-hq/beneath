package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
)

func TestAddMetrics(t *testing.T) {
	m1 := &gql.Metrics{
		ReadOps:      1,
		ReadBytes:    10,
		ReadRecords:  100,
		WriteOps:     1000,
		WriteBytes:   10000,
		WriteRecords: 100000,
	}

	m2 := &gql.Metrics{
		ReadOps:      1,
		ReadBytes:    2,
		ReadRecords:  3,
		WriteOps:     4,
		WriteBytes:   5,
		WriteRecords: 6,
	}

	m := addMetrics(m1, m2)
	assert.Equal(t, m, m1)
	assert.Equal(t, m.ReadOps, 2)
	assert.Equal(t, m.ReadBytes, 12)
	assert.Equal(t, m.ReadRecords, 103)
	assert.Equal(t, m.WriteOps, 1004)
	assert.Equal(t, m.WriteBytes, 10005)
	assert.Equal(t, m.WriteRecords, 100006)
}

func TestMergeMetrics(t *testing.T) {
	xs := []*gql.Metrics{
		makeMetrics(2001, 1),
		makeMetrics(2002, 2),
		makeMetrics(2003, 3),
		makeMetrics(2005, 4),
	}

	ys := []*gql.Metrics{
		makeMetrics(2000, 10),
		makeMetrics(2001, 20),
		makeMetrics(2002, 30),
		makeMetrics(2004, 40),
		makeMetrics(2005, 50),
		makeMetrics(2006, 60),
	}

	zs := mergeUsage(xs, ys)

	assert.Equal(t, len(zs), 7)
	assert.Equal(t, zs[0].Time.Year(), 2000)
	assert.Equal(t, zs[0].ReadOps, 10)
	assert.Equal(t, zs[1].Time.Year(), 2001)
	assert.Equal(t, zs[1].ReadOps, 21)
	assert.Equal(t, zs[2].Time.Year(), 2002)
	assert.Equal(t, zs[2].ReadOps, 32)
	assert.Equal(t, zs[3].Time.Year(), 2003)
	assert.Equal(t, zs[3].ReadOps, 3)
	assert.Equal(t, zs[4].Time.Year(), 2004)
	assert.Equal(t, zs[4].ReadOps, 40)
	assert.Equal(t, zs[5].Time.Year(), 2005)
	assert.Equal(t, zs[5].ReadOps, 54)
	assert.Equal(t, zs[6].Time.Year(), 2006)
	assert.Equal(t, zs[6].ReadOps, 60)
}

func makeMetrics(year, val int) *gql.Metrics {
	return &gql.Metrics{
		Time:         time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		ReadOps:      val,
		ReadBytes:    val,
		ReadRecords:  val,
		WriteOps:     val,
		WriteBytes:   val,
		WriteRecords: val,
	}
}

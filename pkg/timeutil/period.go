package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// Period represents a unit of time
type Period int

const (
	// PeriodNone represents no period
	PeriodNone Period = iota

	// PeriodYear represents years as a Period
	PeriodYear

	// PeriodMonth represents months as a Period
	PeriodMonth

	// PeriodDay represents days as a Period
	PeriodDay

	// PeriodHour represents hours as a Period
	PeriodHour

	// PeriodMinute represents minutes as a Period
	PeriodMinute
)

// PeriodFromString does a best attempt at parsing string as a period
func PeriodFromString(s string) (Period, error) {
	if len(s) == 1 {
		return PeriodFromByte(s[0])
	}

	s = strings.ToLower(s)
	switch s {
	case "minute":
		return PeriodMinute, nil
	case "hour":
		return PeriodHour, nil
	case "day":
		return PeriodDay, nil
	case "month":
		return PeriodMonth, nil
	case "year":
		return PeriodYear, nil
	default:
		return 0, fmt.Errorf("no period matches '%v'", s)
	}
}

// PeriodFromByte is the reverse of Period.Byte()
func PeriodFromByte(b byte) (Period, error) {
	switch b {
	case 'm':
		return PeriodMinute, nil
	case 'H':
		return PeriodHour, nil
	case 'd':
		return PeriodDay, nil
	case 'M':
		return PeriodMonth, nil
	case 'y':
		return PeriodYear, nil
	default:
		return 0, fmt.Errorf("no period matches '%v'", b)
	}
}

// Byte returns the Period as a letter
func (p Period) Byte() byte {
	switch p {
	case PeriodMinute:
		return 'm'
	case PeriodHour:
		return 'H'
	case PeriodDay:
		return 'd'
	case PeriodMonth:
		return 'M'
	case PeriodYear:
		return 'y'
	default:
		panic(fmt.Errorf("unknown period '%d'", p))
	}
}

// String returns the Period as a letter
func (p Period) String() string {
	switch p {
	case PeriodMinute:
		return "Minute"
	case PeriodHour:
		return "Hour"
	case PeriodDay:
		return "Day"
	case PeriodMonth:
		return "Month"
	case PeriodYear:
		return "Year"
	default:
		panic(fmt.Errorf("unknown period '%d'", p))
	}
}

// Count returns the number of periods between two times
func (p Period) Count(from, until time.Time) int {
	switch p {
	case PeriodMinute:
		return int(until.Sub(from).Minutes())
	case PeriodHour:
		return int(until.Sub(from).Hours())
	case PeriodDay:
		return int(until.Sub(from).Hours() / 24)
	case PeriodMonth:
		uy, um, _ := until.Date()
		fy, fm, _ := from.Date()
		count := (uy-fy)*12 + int(um) - int(fm)
		return count
	case PeriodYear:
		uy, _, _ := until.Date()
		fy, _, _ := from.Date()
		count := uy - fy
		return count
	default:
		panic(fmt.Errorf("unknown period '%d'", p))
	}
}

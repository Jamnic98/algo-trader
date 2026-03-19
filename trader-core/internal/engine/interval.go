package engine

import (
	"fmt"
	"time"
)

type Interval int

const (
	Interval1m Interval = iota
	// Interval3m
	Interval5m
	Interval15m
	// Interval30m
	Interval1h
	// Interval2h
	// Interval4h
	// Interval6h
	// Interval8h
	// Interval12h
	Interval1d
)

func (i Interval) Duration() time.Duration {
	switch i {
	case Interval1m:
		return time.Minute

	// case Interval3m:
	// 	return 3 * time.Minute

	case Interval5m:
		return 5 * time.Minute

	case Interval15m:
		return 15 * time.Minute

	// case Interval30m:
	// 	return 30 * time.Minute

	case Interval1h:
		return time.Hour

	// case Interval2h:
	// 	return 2 * time.Hour

	// case Interval4h:
	// 	return 4 * time.Hour

	// case Interval6h:
	// 	return 6 * time.Hour

	// case Interval8h:
	// 	return 8 * time.Hour

	// case Interval12h:
	// 	return 12 * time.Hour

	case Interval1d:
		return 24 * time.Hour

	// case Interval3d:
	// 	return 3 * 24 * time.Hour

	default:
		panic("unknown interval")
	}
}

func (i Interval) String() string {
	switch i {
	case Interval1m:
		return "1m"
	// case Interval3m:
	// 	return "3m"
	case Interval5m:
		return "5m"
	case Interval15m:
		return "15m"
	// case Interval30m:
	// 	return "30m"
	case Interval1h:
		return "1h"
	// case Interval2h:
	// 	return "2h"
	// case Interval4h:
	// 	return "4h"
	// case Interval6h:
	// 	return "6h"
	// case Interval8h:
	// 	return "8h"
	// case Interval12h:
	// 	return "12h"
	case Interval1d:
		return "1d"
	// case Interval3d:
	// 	return "3d"
	default:
		return "unknown"
	}
}

func ParseInterval(s string) (Interval, error) {
	switch s {
	case "1m":
		return Interval1m, nil
	// case "3m":
	// 	return Interval3m, nil
	case "5m":
		return Interval5m, nil
	case "15m":
		return Interval15m, nil
	// case "30m":
	// 	return Interval15m, nil
	case "1h":
		return Interval1h, nil
	// case "2h":
	// 	return Interval2h, nil
	// case "4h":
	// 	return Interval4h, nil
	// case "6h":
	// 	return Interval6h, nil
	// case "8h":
	// 	return Interval8h, nil
	// case "12h":
	// 	return Interval12h, nil
	case "1d":
		return Interval1d, nil
	// case "3d":
	// 	return Interval3d, nil
	default:
		return 0, fmt.Errorf("unsupported interval: %s", s)
	}
}

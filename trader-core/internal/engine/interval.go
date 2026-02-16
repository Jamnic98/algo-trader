package engine

import (
	"fmt"
	"time"
)

type Interval int

const (
	Interval1m Interval = iota
	Interval5m
	Interval15m
	Interval1h
	Interval4h
	Interval1d
)

func (i Interval) Duration() time.Duration {
	switch i {
	case Interval1m:
		return time.Minute
	case Interval5m:
		return 5 * time.Minute
	case Interval15m:
		return 15 * time.Minute
	case Interval1h:
		return time.Hour
	case Interval4h:
		return 4 * time.Hour
	case Interval1d:
		return 24 * time.Hour
	default:
		panic("unknown interval")
	}
}

func (i Interval) String() string {
	switch i {
	case Interval1m:
		return "1m"
	case Interval5m:
		return "5m"
	case Interval15m:
		return "15m"
	case Interval1h:
		return "1h"
	case Interval4h:
		return "4h"
	case Interval1d:
		return "1d"
	default:
		return "unknown"
	}
}

func ParseInterval(s string) (Interval, error) {
	switch s {
	case "1m":
		return Interval1m, nil
	case "5m":
		return Interval5m, nil
	case "15m":
		return Interval15m, nil
	case "1h":
		return Interval1h, nil
	case "4h":
		return Interval4h, nil
	case "1d":
		return Interval1d, nil
	default:
		return 0, fmt.Errorf("unsupported interval: %s", s)
	}
}

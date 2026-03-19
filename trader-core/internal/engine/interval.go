package engine

import (
	"fmt"
	"time"
)

type Interval int

// Intervals in minutes
const (
	Interval1m  Interval = 1
	Interval3m  Interval = 3
	Interval5m  Interval = 5
	Interval15m Interval = 15
	Interval30m Interval = 30
	Interval1h  Interval = 60
	Interval2h  Interval = 120
	Interval4h  Interval = 240
	Interval6h  Interval = 360
	Interval8h  Interval = 480
	Interval12h Interval = 720
	Interval1d  Interval = 1440
	Interval3d  Interval = 4320
	Interval1w  Interval = 10080
	Interval1M  Interval = 43800
)

func (i Interval) Duration() time.Duration {
	return time.Duration(i) * time.Minute
}

func (i Interval) Lookback(candles int) time.Duration {
	return time.Duration(candles) * i.Duration()
}

func (i Interval) String() string {
	switch i {
	case Interval1m:
		return "1m"
	case Interval3m:
		return "3m"
	case Interval5m:
		return "5m"
	case Interval15m:
		return "15m"
	case Interval30m:
		return "30m"
	case Interval1h:
		return "1h"
	case Interval2h:
		return "2h"
	case Interval4h:
		return "4h"
	case Interval6h:
		return "6h"
	case Interval8h:
		return "8h"
	case Interval12h:
		return "12h"
	case Interval1d:
		return "1d"
	case Interval3d:
		return "3d"
	case Interval1w:
		return "1w"
	case Interval1M:
		return "1M"
	default:
		return "unknown"
	}
}

func ParseInterval(s string) (Interval, error) {
	switch s {
	case "1m":
		return Interval1m, nil
	case "3m":
		return Interval3m, nil
	case "5m":
		return Interval5m, nil
	case "15m":
		return Interval15m, nil
	case "30m":
		return Interval30m, nil // was wrongly returning Interval15m
	case "1h":
		return Interval1h, nil
	case "2h":
		return Interval2h, nil
	case "4h":
		return Interval4h, nil
	case "6h":
		return Interval6h, nil
	case "8h":
		return Interval8h, nil
	case "12h":
		return Interval12h, nil
	case "1d":
		return Interval1d, nil
	case "3d":
		return Interval3d, nil
	case "1w":
		return Interval1w, nil
	case "1M":
		return Interval1M, nil
	default:
		return 0, fmt.Errorf("unsupported interval: %s", s)
	}
}

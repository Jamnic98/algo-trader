package strategies

import (
	"fmt"
	"trader-core/internal/db/models"
)

type Signal string

const (
	Hold Signal = "HOLD"
	Buy  Signal = "BUY"
	Sell Signal = "SELL"
)

// Strategy interface for any strategy
type Strategy interface {
	OnCandle(c models.Candle) Signal
}

func NewStrategy(name string) (Strategy, error) {
	switch name {
	case "simple":
		return NewSimpleStrategy(), nil
	// add more here as you build them
	default:
		return nil, fmt.Errorf("unknown strategy: %q", name)
	}
}

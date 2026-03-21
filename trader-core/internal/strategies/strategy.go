package strategies

import (
	"fmt"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type Signal string

const (
	Hold Signal = "HOLD"
	Buy  Signal = "BUY"
	Sell Signal = "SELL"
)

type Decision struct {
	Signal   Signal
	Quantity *decimal.Decimal // nil = caller decides, non-nil = strategy dictates
}

// Strategy interface for any strategy
type Strategy interface {
	OnCandle(ctx CandleContext) Decision
}

func New(name string) (Strategy, error) {
	switch name {
	case "simple":
		return NewSimpleStrategy(), nil
	case "simpleDca":
		return NewSimpleDCAStrategy(), nil
	// case "smartDca":
	// 	return NewSmartDCAStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %q", name)
	}
}

type CandleContext struct {
	Candle   models.Candle
	FeeRate  decimal.Decimal
	AvgEntry decimal.Decimal // 0 if no position
}

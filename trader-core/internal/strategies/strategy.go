package strategies

import (
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type Signal string

type Logger interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

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

type CandleContext struct {
	Candle   models.Candle
	FeeRate  decimal.Decimal
	AvgEntry decimal.Decimal // 0 if no position
}

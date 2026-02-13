package strategies

import "trader-core/internal/db/models"

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

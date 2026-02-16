package strategies

import (
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
)

type SimpleStrategy struct {
	hasPosition bool
}

func NewSimpleStrategy() *SimpleStrategy {
	return &SimpleStrategy{}
}

func (s SimpleStrategy) OnCandles(candles []models.Candle) engine.Side {
	candlesLen := len(candles)
	last := candles[candlesLen-1]
	prev := candles[candlesLen-2]

	if last.Close > prev.Close {
		return engine.BUY
	}

	if last.Close < prev.Close {
		return engine.SELL
	}

	return engine.NONE
}

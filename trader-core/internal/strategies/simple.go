package strategies

import (
	"trader-core/internal/db/models"
)

type SimpleStrategy struct {
	hasPosition bool
	prev        *models.Candle
}

func NewSimpleStrategy() *SimpleStrategy {
	return &SimpleStrategy{}
}

func (s *SimpleStrategy) OnCandle(c models.Candle) Signal {
	defer func() { s.prev = &c }()

	if s.prev == nil {
		return Hold // not enough data yet
	}

	if !s.hasPosition && c.Close.LessThan(s.prev.Close) {
		s.hasPosition = true
		return Buy
	}

	if s.hasPosition && c.Close.GreaterThan(s.prev.Close) {
		s.hasPosition = false
		return Sell
	}

	return Hold
}

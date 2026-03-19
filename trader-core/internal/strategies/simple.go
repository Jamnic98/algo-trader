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

func (s *SimpleStrategy) OnCandle(c models.Candle) Decision {
	defer func() { s.prev = &c }()

	if s.prev == nil {
		return Decision{Signal: Hold}
	}

	if !s.hasPosition && c.Close.LessThan(s.prev.Close) {
		s.hasPosition = true
		return Decision{Signal: Buy} // Quantity: nil — caller decides
	}

	if s.hasPosition && c.Close.GreaterThan(s.prev.Close) {
		s.hasPosition = false
		return Decision{Signal: Sell} // Quantity: nil — sell all, caller decides
	}

	return Decision{Signal: Hold}
}

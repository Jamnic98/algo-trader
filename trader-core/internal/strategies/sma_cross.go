package strategies

import (
	"github.com/shopspring/decimal"
)

type SMACrossParams struct {
	Fast int `json:"fast"`
	Slow int `json:"slow"`
}

// SMACrossStrategy implements the "sma_cross" template: buy when the fast SMA
// crosses above the slow SMA, sell when it crosses below. Mirrored by
// research/app/strategies/sma_cross.py; behavior pinned by the shared test
// vectors in shared/strategies.json.
type SMACrossStrategy struct {
	fast        int
	slow        int
	closes      []decimal.Decimal
	hasPosition bool
}

func NewSMACrossStrategy(params SMACrossParams) *SMACrossStrategy {
	if params.Fast < 1 {
		params.Fast = 5
	}
	if params.Slow < 2 {
		params.Slow = 10
	}
	if params.Fast >= params.Slow {
		params.Fast = params.Slow - 1
	}
	return &SMACrossStrategy{fast: params.Fast, slow: params.Slow}
}

func mean(values []decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

func (s *SMACrossStrategy) OnCandle(ctx CandleContext) Decision {
	s.closes = append(s.closes, ctx.Candle.Close)
	if len(s.closes) > s.slow {
		s.closes = s.closes[len(s.closes)-s.slow:]
	}
	if len(s.closes) < s.slow {
		return Decision{Signal: Hold}
	}

	fast := mean(s.closes[len(s.closes)-s.fast:])
	slow := mean(s.closes)

	if !s.hasPosition && fast.GreaterThan(slow) {
		s.hasPosition = true
		return Decision{Signal: Buy}
	}
	if s.hasPosition && fast.LessThan(slow) {
		s.hasPosition = false
		return Decision{Signal: Sell}
	}
	return Decision{Signal: Hold}
}

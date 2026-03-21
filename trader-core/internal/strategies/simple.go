package strategies

import (
	"fmt"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type SimpleParams struct{}
type SimpleStrategy struct {
	hasPosition bool
	prev        *models.Candle
	buyPrice    decimal.Decimal
	logger      Logger
}

func NewSimpleStrategy(params SimpleParams, logger Logger) *SimpleStrategy {
	return &SimpleStrategy{logger: logger}
}

func (s *SimpleStrategy) OnCandle(ctx CandleContext) Decision {
	c := ctx.Candle
	defer func() { s.prev = &c }()

	if s.prev == nil {
		return Decision{Signal: Hold}
	}

	// --- BUY ---
	if !s.hasPosition && c.Close.LessThan(s.prev.Close) {
		// no fee check here — should verify round trip is coverable
		s.hasPosition = true
		s.buyPrice = c.Close
		return Decision{Signal: Buy}
	}

	// --- SELL ---
	if s.hasPosition && c.Close.GreaterThan(s.prev.Close) {
		roundTripFee := s.buyPrice.Mul(ctx.FeeRate.Mul(decimal.NewFromInt(2)))
		priceDiff := c.Close.Sub(s.buyPrice)

		if priceDiff.LessThanOrEqual(roundTripFee) {
			reason := fmt.Sprintf("fee gate: move %s < required %s",
				priceDiff.String(),
				roundTripFee.String(),
			)
			return Decision{Signal: Hold, Reason: reason}
		}

		s.hasPosition = false
		s.buyPrice = decimal.Zero
		return Decision{Signal: Sell}
	}

	return Decision{Signal: Hold}
}

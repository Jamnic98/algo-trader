package strategies

import (
	"trader-core/internal/common"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type SimpleParams struct{}
type SimpleStrategy struct {
	hasPosition bool
	prev        *models.Candle
	buyPrice    decimal.Decimal
	fees        common.FeeConfig
	logger      Logger
}

func NewSimpleStrategy(params SimpleParams, logger Logger) *SimpleStrategy {
	return &SimpleStrategy{
		logger: logger,
	}
}

func (s *SimpleStrategy) OnCandle(ctx CandleContext) Decision {
	c := ctx.Candle
	defer func() { s.prev = &c }()

	if s.prev == nil {
		return Decision{Signal: Hold}
	}

	// --- BUY ---
	if !s.hasPosition && c.Close.LessThan(s.prev.Close) {
		s.hasPosition = true
		s.buyPrice = c.Close
		return Decision{Signal: Buy}
	}

	// --- SELL ---
	if s.hasPosition && c.Close.GreaterThan(s.prev.Close) {
		minMove := s.fees.MinPriceMovement(s.buyPrice)
		priceDiff := c.Close.Sub(s.buyPrice)

		if priceDiff.LessThanOrEqual(minMove) {
			s.logger.Info(
				"[SimpleStrategy] hold — price move %s does not cover fees %s (buy: %s, now: %s)",
				priceDiff.StringFixed(8),
				minMove.StringFixed(8),
				s.buyPrice.StringFixed(8),
				c.Close.StringFixed(8),
			)
			return Decision{Signal: Hold}
		}

		s.hasPosition = false
		s.buyPrice = decimal.Zero
		return Decision{Signal: Sell}
	}

	return Decision{Signal: Hold}
}

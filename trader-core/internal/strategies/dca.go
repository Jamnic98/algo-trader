package strategies

import (
	"github.com/shopspring/decimal"
)

type SimpleDCAStrategy struct{}
type SimpleDCAParams struct{}

func NewSimpleDCAStrategy(params SimpleDCAParams, logger Logger) *SimpleDCAStrategy {
	return &SimpleDCAStrategy{}
}

func (s *SimpleDCAStrategy) OnCandle(ctx CandleContext) Decision {
	return Decision{Signal: Buy}
}

type SmartDCAParams struct {
	FixedSpend string `json:"fixed_spend"` // string → decimal, e.g. "50.00"
}
type SmartDCAStrategy struct {
	fixedSpend  decimal.Decimal
	totalSpent  decimal.Decimal
	totalBought decimal.Decimal
	logger      Logger
}

// TODO: implement
func NewSmartDCAStrategy(params SmartDCAParams, logger Logger) *SmartDCAStrategy {

	fixedSpend, err := decimal.NewFromString(params.FixedSpend)
	if err != nil || fixedSpend.IsZero() {
		logger.Info("[SmartDCA] invalid fixed_spend %q, defaulting to 10", params.FixedSpend)
		fixedSpend = decimal.NewFromInt(10)
	}
	return &SmartDCAStrategy{
		fixedSpend: fixedSpend,
		// totalSpent:  s.totalSpent,
		// totalBought: s.totalBought,
	}
}

func (s *SmartDCAStrategy) OnCandle(ctx CandleContext) Decision {
	c := ctx.Candle
	price := c.Close

	if s.totalBought.IsPositive() {
		avgEntry := s.totalSpent.Div(s.totalBought)
		threshold := avgEntry.Mul(decimal.NewFromFloat(1.05))

		if price.GreaterThan(threshold) {
			// check sell side fee is covered
			sellValue := price.Mul(s.totalBought)
			sellFee := sellValue.Mul(ctx.FeeRate)
			profit := sellValue.Sub(s.totalSpent).Sub(sellFee)

			if profit.IsNegative() {
				s.logger.Info(
					"[SmartDCA] skipping sell — profit %s negative after fee %s",
					profit.StringFixed(8),
					sellFee.StringFixed(8),
				)
				return Decision{Signal: Hold}
			}

			qtyToSell := s.totalBought
			s.totalSpent = decimal.Zero
			s.totalBought = decimal.Zero
			return Decision{Signal: Sell, Quantity: &qtyToSell}
		}
	}

	// check buy side fee is covered by fixed spend
	qtyToBuy := s.calculateQty(price)
	cost := price.Mul(qtyToBuy)
	buyFee := cost.Mul(ctx.FeeRate)

	if buyFee.GreaterThanOrEqual(s.fixedSpend) {
		s.logger.Info(
			"[SmartDCA] skipping buy — fee %s exceeds fixed spend %s",
			buyFee.StringFixed(8),
			s.fixedSpend.StringFixed(8),
		)
		return Decision{Signal: Hold}
	}

	s.totalSpent = s.totalSpent.Add(cost.Add(buyFee)) // include fee in cost basis
	s.totalBought = s.totalBought.Add(qtyToBuy)
	return Decision{Signal: Buy, Quantity: &qtyToBuy}
}

func (s *SmartDCAStrategy) calculateQty(price decimal.Decimal) decimal.Decimal {
	return s.fixedSpend.Div(price)
}

package strategies

import (
	"github.com/shopspring/decimal"
)

// registration happens in definitions.go, driven by shared/strategies.json

type SimpleDCAStrategy struct{}
type SimpleDCAParams struct{}

func NewSimpleDCAStrategy(params SimpleDCAParams) *SimpleDCAStrategy {
	return &SimpleDCAStrategy{}
}

func (s *SimpleDCAStrategy) OnCandle(ctx CandleContext) Decision {
	return Decision{Signal: Buy}
}

type SmartDCAParams struct {
	FixedSpend string  `json:"fixed_spend"` // string → decimal, e.g. "50.00"
	TakeProfit float64 `json:"take_profit"` // multiplier over avg entry, e.g. 1.05
}

type SmartDCAStrategy struct {
	fixedSpend  decimal.Decimal
	takeProfit  decimal.Decimal
	totalSpent  decimal.Decimal
	totalBought decimal.Decimal
}

func NewSmartDCAStrategy(params SmartDCAParams) *SmartDCAStrategy {
	fixedSpend, err := decimal.NewFromString(params.FixedSpend)
	if err != nil || fixedSpend.IsZero() {
		fixedSpend = decimal.NewFromInt(10)
	}
	takeProfit := params.TakeProfit
	if takeProfit <= 0 {
		takeProfit = 1.05
	}
	return &SmartDCAStrategy{
		fixedSpend: fixedSpend,
		takeProfit: decimal.NewFromFloat(takeProfit),
	}
}

func (s *SmartDCAStrategy) OnCandle(ctx CandleContext) Decision {
	c := ctx.Candle
	price := c.Close

	if s.totalBought.IsPositive() {
		avgEntry := s.totalSpent.Div(s.totalBought)
		threshold := avgEntry.Mul(s.takeProfit)

		if price.GreaterThan(threshold) {
			// check sell side fee is covered
			sellValue := price.Mul(s.totalBought)
			sellFee := sellValue.Mul(ctx.FeeRate)
			profit := sellValue.Sub(s.totalSpent).Sub(sellFee)

			if profit.IsNegative() {
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
		return Decision{Signal: Hold}
	}

	s.totalSpent = s.totalSpent.Add(cost.Add(buyFee)) // include fee in cost basis
	s.totalBought = s.totalBought.Add(qtyToBuy)
	return Decision{Signal: Buy, Quantity: &qtyToBuy}
}

func (s *SmartDCAStrategy) calculateQty(price decimal.Decimal) decimal.Decimal {
	return s.fixedSpend.Div(price)
}

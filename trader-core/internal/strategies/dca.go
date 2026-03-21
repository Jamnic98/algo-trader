package strategies

import (
	"github.com/shopspring/decimal"
)

type SimpleDCAStrategy struct{}

func NewSimpleDCAStrategy() *SimpleDCAStrategy {
	return &SimpleDCAStrategy{}
}

func (s *SimpleDCAStrategy) OnCandle(ctx CandleContext) Decision {
	return Decision{Signal: Buy}
}

type SmartDCAStrategy struct {
	fixedSpend  decimal.Decimal
	totalSpent  decimal.Decimal
	totalBought decimal.Decimal
}

// TODO: implement
func NewSmartDCAStrategy(s SmartDCAStrategy) *SmartDCAStrategy {
	return &SmartDCAStrategy{
		fixedSpend: s.fixedSpend,
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
			qtyToSell := s.totalBought
			s.totalSpent = decimal.Zero
			s.totalBought = decimal.Zero
			return Decision{Signal: Sell, Quantity: &qtyToSell}
		}
	}

	qtyToBuy := s.calculateQty(price)
	s.totalSpent = s.totalSpent.Add(price.Mul(qtyToBuy))
	s.totalBought = s.totalBought.Add(qtyToBuy)
	return Decision{Signal: Buy, Quantity: &qtyToBuy}
}

func (s *SmartDCAStrategy) calculateQty(price decimal.Decimal) decimal.Decimal {
	return s.fixedSpend.Div(price)
}

package strategies

import (
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type SimpleDCAStrategy struct{}

func NewSimpleDCAStrategy() *SimpleDCAStrategy {
	return &SimpleDCAStrategy{}
}

func (s *SimpleDCAStrategy) OnCandle(c models.Candle) Decision {
	return Decision{Signal: Buy}
}

type SmartDCAStrategy struct {
	fixedSpend  decimal.Decimal
	totalSpent  decimal.Decimal
	totalBought decimal.Decimal
}

func NewSmartDCAStrategy(fixedSpend decimal.Decimal) *SmartDCAStrategy {
	return &SmartDCAStrategy{fixedSpend: fixedSpend}
}

func (s *SmartDCAStrategy) OnCandle(c models.Candle) Decision {
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

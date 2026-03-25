package strategies

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

func init() {
	Register("simple_dca", StrategySchema{
		Name:        "simple_dca",
		DisplayName: "Simple DCA",
		Params:      []ParamSchema{},
	}, func(_ []byte) (Strategy, error) {
		return NewSimpleDCAStrategy(SimpleDCAParams{}), nil
	})

	minSpend := 0.01
	Register("smart_dca", StrategySchema{
		Name:        "smart_dca",
		DisplayName: "Smart DCA",
		Params: []ParamSchema{
			{
				Key:     "fixed_spend",
				Label:   "Fixed Spend",
				Type:    ParamTypeString,
				Default: "10.00",
				Min:     &minSpend,
			},
		},
	}, func(paramsJSON []byte) (Strategy, error) {
		var params SmartDCAParams
		if len(paramsJSON) > 0 {
			if err := json.Unmarshal(paramsJSON, &params); err != nil {
				return nil, fmt.Errorf("parsing smart_dca params: %w", err)
			}
		}
		if params.FixedSpend == "" {
			params.FixedSpend = "10.00"
		}
		// logger passed as nil here — bot factory will need to inject it
		return NewSmartDCAStrategy(params), nil
	})
}

type SimpleDCAStrategy struct{}
type SimpleDCAParams struct{}

func NewSimpleDCAStrategy(params SimpleDCAParams) *SimpleDCAStrategy {
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
}

func NewSmartDCAStrategy(params SmartDCAParams) *SmartDCAStrategy {
	fixedSpend, err := decimal.NewFromString(params.FixedSpend)
	if err != nil || fixedSpend.IsZero() {
		fixedSpend = decimal.NewFromInt(10)
	}
	return &SmartDCAStrategy{fixedSpend: fixedSpend}
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

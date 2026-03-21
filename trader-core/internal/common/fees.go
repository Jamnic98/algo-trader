package common

import "github.com/shopspring/decimal"

type FeeConfig struct {
	MakerFee decimal.Decimal // e.g. 0.001 for 0.1%
	TakerFee decimal.Decimal // e.g. 0.001 for 0.1%
}

// SpotFeeConfig returns standard Binance spot fees
func SpotFeeConfig() FeeConfig {
	return FeeConfig{
		MakerFee: decimal.NewFromFloat(0.001),
		TakerFee: decimal.NewFromFloat(0.001),
	}
}

// SpotBNBFeeConfig returns reduced Binance spot fees when paying with BNB
func SpotBNBFeeConfig() FeeConfig {
	return FeeConfig{
		MakerFee: decimal.NewFromFloat(0.00075),
		TakerFee: decimal.NewFromFloat(0.00075),
	}
}

// RoundTripCost returns the total fee cost for a full buy+sell cycle
// e.g. at 0.1% each side: price * 0.002
func (f FeeConfig) RoundTripCost(price decimal.Decimal) decimal.Decimal {
	return price.Mul(f.TakerFee.Add(f.MakerFee))
}

// MinPriceMovement returns the minimum price increase needed to cover round trip fees
func (f FeeConfig) MinPriceMovement(buyPrice decimal.Decimal) decimal.Decimal {
	return f.RoundTripCost(buyPrice)
}

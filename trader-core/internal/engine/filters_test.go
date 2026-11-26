package engine

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateOrder(t *testing.T) {
	f := SymbolFilters{
		MinQty:      decimal.NewFromFloat(0.001),
		StepQty:     decimal.NewFromFloat(0.001),
		TickSize:    decimal.NewFromFloat(0.01),
		MinNotional: decimal.NewFromInt(10),
	}
	ok := Order{Symbol: "BTCUSDT", Signal: "BUY", Price: decimal.NewFromInt(50000), Qty: decimal.NewFromFloat(0.002)}
	if err := ValidateOrder(f, ok); err != nil {
		t.Fatalf("valid order rejected: %v", err)
	}

	cases := []struct {
		name string
		ord  Order
	}{
		{"below min qty", Order{Price: decimal.NewFromInt(50000), Qty: decimal.NewFromFloat(0.0005)}},
		{"off step qty", Order{Price: decimal.NewFromInt(50000), Qty: decimal.NewFromFloat(0.0015)}},
		{"below min notional", Order{Price: decimal.NewFromInt(100), Qty: decimal.NewFromFloat(0.05)}}, // 5 < 10
		{"off tick price", Order{Price: decimal.NewFromFloat(50000.005), Qty: decimal.NewFromInt(1)}},
	}
	for _, c := range cases {
		if err := ValidateOrder(f, c.ord); err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
		}
	}

	// zero filters = no constraints
	if err := ValidateOrder(SymbolFilters{}, Order{Price: decimal.NewFromInt(1), Qty: decimal.NewFromFloat(0.0001)}); err != nil {
		t.Fatalf("zero filters should allow anything: %v", err)
	}
}

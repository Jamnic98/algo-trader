package engine

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// SymbolFilters are the venue's trading rules for a symbol (lot size, tick
// size, min notional). Zero value = no constraint.
type SymbolFilters struct {
	MinQty      decimal.Decimal
	MaxQty      decimal.Decimal
	StepQty     decimal.Decimal
	TickSize    decimal.Decimal
	StepPrice   decimal.Decimal
	MinNotional decimal.Decimal
}

// ValidateOrder checks an order against the venue filters. Returns nil if OK.
func ValidateOrder(f SymbolFilters, o Order) error {
	if f.MinQty.GreaterThan(decimal.Zero) && o.Qty.LessThan(f.MinQty) {
		return fmt.Errorf("qty %s below min %s", o.Qty.String(), f.MinQty.String())
	}
	if f.MaxQty.GreaterThan(decimal.Zero) && o.Qty.GreaterThan(f.MaxQty) {
		return fmt.Errorf("qty %s above max %s", o.Qty.String(), f.MaxQty.String())
	}
	if f.StepQty.GreaterThan(decimal.Zero) {
		rem := o.Qty.Mod(f.StepQty)
		if !rem.IsZero() {
			return fmt.Errorf("qty %s not a multiple of step %s", o.Qty.String(), f.StepQty.String())
		}
	}
	if f.MinNotional.GreaterThan(decimal.Zero) {
		notional := o.Price.Mul(o.Qty)
		if notional.LessThan(f.MinNotional) {
			return fmt.Errorf("notional %s below min %s", notional.String(), f.MinNotional.String())
		}
	}
	if f.TickSize.GreaterThan(decimal.Zero) && o.Price.GreaterThan(decimal.Zero) {
		rem := o.Price.Mod(f.TickSize)
		if !rem.IsZero() {
			return fmt.Errorf("price %s not a multiple of tick %s", o.Price.String(), f.TickSize.String())
		}
	}
	return nil
}

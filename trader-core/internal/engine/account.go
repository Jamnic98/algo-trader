package engine

import (
	"github.com/shopspring/decimal"
)

type AccountSnapshot struct {
	Balance    decimal.Decimal
	Positions  map[string]decimal.Decimal
	CostBasis  map[string]decimal.Decimal // symbol -> total quote spent (for avg entry / pnl)
}

type Account interface {
	ApplyFill(*Fill) error
	Snapshot() AccountSnapshot
}

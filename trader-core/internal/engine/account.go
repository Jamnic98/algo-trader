package engine

import (
	"github.com/shopspring/decimal"
)

type AccountSnapshot struct {
	Balance   decimal.Decimal
	Positions map[string]decimal.Decimal
}

type Account interface {
	ApplyFill(*Fill) error
	Snapshot() AccountSnapshot
}

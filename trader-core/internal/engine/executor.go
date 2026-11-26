package engine

import "github.com/shopspring/decimal"

type ExecutionEngine interface {
	ExecuteFill(Order) (*Fill, error)
}

// FeeRater is implemented by executors that know their account's fee rate,
// so strategies see the real fee instead of a hardcoded default.
type FeeRater interface {
	FeeRate() decimal.Decimal
}

// AccountSnapshotter is implemented by executors that can expose their
// account ledger (balance, positions, cost basis) — used by risk management.
type AccountSnapshotter interface {
	Snapshot() AccountSnapshot
}

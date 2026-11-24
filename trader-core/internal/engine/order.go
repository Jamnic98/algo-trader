package engine

import (
	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

type Order struct {
	BotID  string
	Symbol string
	Signal strategies.Signal
	Price  decimal.Decimal
	Qty    decimal.Decimal
}

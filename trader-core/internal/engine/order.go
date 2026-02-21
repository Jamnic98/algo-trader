package engine

import "github.com/shopspring/decimal"

type Order struct {
	BotID  string
	Symbol string
	Side   Side
	Price  decimal.Decimal
	Qty    decimal.Decimal
}

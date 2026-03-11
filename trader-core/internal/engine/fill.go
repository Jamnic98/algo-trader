package engine

import (
	"time"

	"github.com/shopspring/decimal"
)

type Fill struct {
	BotID    string          // which bot submitted the order
	Symbol   string          // trading pair, e.g., BTCUSDT
	Qty      decimal.Decimal // positive = buy, negative = sell
	Price    decimal.Decimal // execution price
	Notional decimal.Decimal // Qty * Price
	Fee      decimal.Decimal
	Side     Side
	Time     time.Time
}

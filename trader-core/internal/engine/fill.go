package engine

import (
	"time"
	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

type Fill struct {
	BotID  string          // which bot submitted the order
	Symbol string          // trading pair, e.g., BTCUSDT
	Qty    decimal.Decimal // positive = buy, negative = sell
	Price  decimal.Decimal // execution price
	Fee    decimal.Decimal
	Signal strategies.Signal
	Time   time.Time
}

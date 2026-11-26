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
	Pnl    decimal.Decimal // realized pnl on SELLs (0 on BUYs); used by risk cooldown
	Signal strategies.Signal
	Time   time.Time
}

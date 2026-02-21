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

type Position struct {
	Symbol   string
	Qty      decimal.Decimal // positive = long, negative = short
	AvgPrice decimal.Decimal // average entry price
}

func (p *Position) Apply(fill Fill) {
	totalQty := p.Qty.Add(fill.Qty)

	zero := decimal.NewFromInt(0)

	if totalQty.Equal(zero) {
		// position closed
		p.Qty = zero
		p.AvgPrice = zero
		return
	}

	if p.Qty.Equal(zero) {
		// opening new position
		p.Qty = fill.Qty
		p.AvgPrice = fill.Price
		return
	}

	// updating existing position with weighted average price
	// (AvgPrice*Qty + fill.Price*fill.Qty) / totalQty
	p.AvgPrice = (p.AvgPrice.Mul(p.Qty).Add(fill.Price.Mul(fill.Qty))).Div(totalQty)
	p.Qty = totalQty
}

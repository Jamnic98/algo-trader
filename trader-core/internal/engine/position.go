package engine

import "time"

type Fill struct {
	BotID  string  // which bot submitted the order
	Symbol string  // trading pair, e.g., BTCUSDT
	Qty    float64 // positive = buy, negative = sell
	Price  float64 // execution price
	Value  float64 // Qty * Price
	Side   Side
	Fee    float64
	Time   time.Time
}

type Position struct {
	Symbol   string
	Qty      float64 // positive = long, negative = short
	AvgPrice float64 // average entry price
}

func (p *Position) Apply(fill Fill) {
	totalQty := p.Qty + fill.Qty

	if totalQty == 0 {
		// position closed
		p.Qty = 0
		p.AvgPrice = 0
		return
	}

	if p.Qty == 0 {
		// opening new position
		p.Qty = fill.Qty
		p.AvgPrice = fill.Price
		return
	}

	// updating existing position
	// weighted average price
	p.AvgPrice = (p.AvgPrice*p.Qty + fill.Price*fill.Qty) / totalQty
	p.Qty = totalQty
}

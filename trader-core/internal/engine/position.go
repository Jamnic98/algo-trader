package engine

import (
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

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

type PositionState struct {
	HasPosition bool
	AvgEntry    decimal.Decimal
	TotalSpent  decimal.Decimal
	TotalBought decimal.Decimal
}

func RestorePosition(db *gorm.DB, botID string) (PositionState, error) {
	var trades []models.Trade
	if err := db.Where("bot_id = ?", botID).Order("timestamp ASC").Find(&trades).Error; err != nil {
		return PositionState{}, err
	}

	scale := decimal.NewFromInt(1e8)
	var totalSpent, totalBought decimal.Decimal

	for _, t := range trades {
		price := decimal.NewFromInt(t.PriceInt).Div(scale)
		qty := decimal.NewFromInt(t.QuantityInt).Div(scale)

		switch t.Side {
		case "BUY":
			totalSpent = totalSpent.Add(price.Mul(qty))
			totalBought = totalBought.Add(qty)
		case "SELL":
			totalSpent = decimal.Zero
			totalBought = decimal.Zero
		}
	}

	hasPosition := totalBought.IsPositive()
	avgEntry := decimal.Zero
	if hasPosition {
		avgEntry = totalSpent.Div(totalBought)
	}

	return PositionState{
		HasPosition: hasPosition,
		AvgEntry:    avgEntry,
		TotalSpent:  totalSpent,
		TotalBought: totalBought,
	}, nil
}

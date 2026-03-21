package dto

import (
	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type PositionDTO struct {
	Symbol     string `json:"symbol"`
	Qty        string `json:"qty"`
	AvgEntry   string `json:"avg_entry"`
	TotalSpent string `json:"total_spent"`
	TotalFees  string `json:"total_fees"`
	Realised   string `json:"realised_pnl"`
}

func CalcBotPositions(botID string) ([]PositionDTO, error) {
	var trades []models.Trade
	if err := db.DB.Where("bot_id = ?", botID).Order("timestamp ASC").Find(&trades).Error; err != nil {
		return nil, err
	}

	type state struct {
		qty         decimal.Decimal
		totalSpent  decimal.Decimal // sum of buy notionals
		totalBought decimal.Decimal // total qty bought
		totalFees   decimal.Decimal
		realised    decimal.Decimal
	}

	ps := make(map[string]*state)

	priceScale := decimal.NewFromInt(1e8)
	qtyScale := decimal.NewFromInt(1e6)
	feeScale := decimal.NewFromInt(1e8)

	for _, t := range trades {
		if _, ok := ps[t.Symbol]; !ok {
			ps[t.Symbol] = &state{}
		}
		s := ps[t.Symbol]

		price := decimal.NewFromInt(t.PriceInt).Div(priceScale)
		qty := decimal.NewFromInt(t.QuantityInt).Div(qtyScale)
		fee := decimal.NewFromInt(t.FeeInt).Div(feeScale)
		notional := price.Mul(qty)

		s.totalFees = s.totalFees.Add(fee)

		switch t.Side {
		case "BUY":
			s.qty = s.qty.Add(qty)
			s.totalSpent = s.totalSpent.Add(notional)
			s.totalBought = s.totalBought.Add(qty)

		case "SELL":
			avgEntry := s.totalSpent.Div(s.totalBought)
			s.realised = s.realised.Add(price.Sub(avgEntry).Mul(qty))

			s.totalSpent = s.totalSpent.Sub(price.Mul(qty))
			s.totalBought = s.totalBought.Sub(qty)

		}
	}

	result := make([]PositionDTO, 0, len(ps))
	for symbol, s := range ps {
		avgEntry := decimal.Zero
		if s.totalBought.IsPositive() {
			avgEntry = s.totalSpent.Div(s.totalBought)
		}
		result = append(result, PositionDTO{
			Symbol:     symbol,
			Qty:        s.qty.StringFixed(6),
			AvgEntry:   avgEntry.StringFixed(8),
			TotalSpent: s.totalSpent.StringFixed(8),
			TotalFees:  s.totalFees.StringFixed(8),
			Realised:   s.realised.StringFixed(8),
		})
	}

	return result, nil
}

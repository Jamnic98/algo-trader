package dto

import (
	"time"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type FIllDTO struct {
	ID        uint            `json:"id"`
	BotID     string          `json:"botID"`
	Symbol    string          `json:"symbol"`
	Base      string          `json:"base"`
	Quote     string          `json:"quote"`
	Side      string          `json:"side"`  // BUY / SELL / NONE
	Price     decimal.Decimal `json:"price"` // per unit price
	Quantity  decimal.Decimal `json:"quantity"`
	Fee       decimal.Decimal `json:"fee"`
	FeeAsset  string          `json:"feeAsset"` // e.g. "USDT", "BTC"
	Exchange  string          `json:"exchange"`
	Timestamp time.Time       `json:"timestamp"`
	CreatedAt time.Time       `json:"createdAt"` // GORM convention
}

func FIllToDTO(t *models.Fill) FIllDTO {
	return FIllDTO{
		ID:        t.ID,
		BotID:     t.BotID,
		Symbol:    t.Symbol,
		Base:      t.Base,
		Quote:     t.Quote,
		Side:      t.Side,
		Price:     decimal.NewFromInt(t.PriceInt).Div(decimal.NewFromInt(priceScale)),
		Quantity:  decimal.NewFromInt(t.QuantityInt).Div(decimal.NewFromInt(quantityScale)),
		Fee:       decimal.NewFromInt(t.FeeInt).Div(decimal.NewFromInt(feeScale)),
		FeeAsset:  t.FeeAsset,
		Exchange:  t.Exchange,
		Timestamp: t.Timestamp,
		CreatedAt: t.CreatedAt,
	}
}

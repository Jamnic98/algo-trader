package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Trade struct {
	ID     uint `gorm:"primaryKey"`
	BotID  string
	Symbol string
	Side   string

	PriceInt    int64 // scaled, e.g. 1e8
	QuantityInt int64 // scaled, e.g. 1e6
	FeeInt      int64 // scaled, e.g. 1e8

	FeeAsset  string
	Exchange  string
	Timestamp time.Time
	CreatedAt time.Time
}

type TradeDTO struct {
	ID        uint            `json:"id"`
	BotID     string          `json:"botID"`
	Symbol    string          `json:"symbol"`
	Side      string          `json:"side"`  // BUY / SELL / NONE
	Price     decimal.Decimal `json:"price"` // per unit price
	Quantity  decimal.Decimal `json:"quantity"`
	Fee       decimal.Decimal `json:"fee"`
	FeeAsset  string          `json:"feeAsset"` // e.g. "USDT", "BTC"
	Exchange  string          `json:"exchange"`
	Timestamp time.Time       `json:"timestamp"`
	CreatedAt time.Time       `json:"createdAt"` // GORM convention
}

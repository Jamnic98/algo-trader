package models

import "time"

type Trade struct {
	ID        uint   `gorm:"primaryKey"`
	BotID     string // which bot/strategy generated the trade
	Symbol    string
	Side      string  // BUY / SELL / NONE
	Price     float64 // per unit price
	Quantity  float64
	Fee       float64 // in fee asset
	FeeAsset  string  // e.g. "USDT", "BTC"
	Exchange  string
	Timestamp time.Time
	CreatedAt time.Time // GORM convention
}

type TradeDTO struct {
	ID        uint      `json:"id"`
	BotID     string    `json:"botID"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`  // BUY / SELL / NONE
	Price     float64   `json:"price"` // per unit price
	Quantity  float64   `json:"quantity"`
	Fee       float64   `json:"fee"`
	FeeAsset  string    `json:"feeAsset"` // e.g. "USDT", "BTC"
	Exchange  string    `json:"exchange"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"createdAt"` // GORM convention
}

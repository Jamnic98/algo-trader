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

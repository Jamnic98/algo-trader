package models

import "time"

type Trade struct {
    ID        uint      `gorm:"primaryKey"`
    Symbol    string
    Side      string
    Price     float64   // per unit price
    Quantity  float64
    Fee       float64   // in fee asset
    FeeAsset  string    // e.g. "USDT", "BTC"
    Exchange  string
    Timestamp time.Time
}

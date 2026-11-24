package models

import (
	"time"
)

type Fill struct {
	ID     uint `gorm:"primaryKey"`
	BotID  string
	Symbol string
	Base   string
	Quote  string
	Side   string

	PriceInt    int64 // scaled, e.g. 1e8
	QuantityInt int64 // scaled, e.g. 1e6
	FeeInt      int64 // scaled, e.g. 1e8

	FeeAsset  string
	Exchange  string
	Timestamp time.Time
	CreatedAt time.Time
}

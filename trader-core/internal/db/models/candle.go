package models

type Candle struct {
	ID        uint `gorm:"primaryKey"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

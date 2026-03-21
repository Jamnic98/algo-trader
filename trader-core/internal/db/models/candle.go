package models

import (
	"github.com/shopspring/decimal"
)

type Candle struct {
	Open      decimal.Decimal `json:"open"`
	Close     decimal.Decimal `json:"close"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Volume    decimal.Decimal `json:"volume"`
	OpenTime  int64           `json:"openTime"`
	CloseTime int64           `json:"closeTime"`
}

type CandleDTO struct {
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    float64 `json:"volume"`
	OpenTime  int64   `json:"openTime"`
	CloseTime int64   `json:"closeTime"`
}

func CandleToDTO(c Candle) CandleDTO {
	return CandleDTO{
		Open:      c.Open.InexactFloat64(),
		Close:     c.Close.InexactFloat64(),
		High:      c.High.InexactFloat64(),
		Low:       c.Low.InexactFloat64(),
		Volume:    c.Volume.InexactFloat64(),
		OpenTime:  c.OpenTime,
		CloseTime: c.CloseTime,
	}
}

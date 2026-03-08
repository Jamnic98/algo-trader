package engine

import (
	"encoding/json"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

type KlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	K         struct {
		StartTime int64       `json:"t"`
		CloseTime int64       `json:"T"`
		Symbol    string      `json:"s"`
		Interval  string      `json:"i"`
		Open      json.Number `json:"o"`
		Close     json.Number `json:"c"`
		High      json.Number `json:"h"`
		Low       json.Number `json:"l"`
		Volume    json.Number `json:"v"`
		Trades    int64       `json:"n"`
		IsClosed  bool        `json:"x"`
	} `json:"k"`
}

// parseKline returns a candle and key only if closed
func ParseKline(msg []byte) (models.Candle, string, bool) {
	var evt KlineEvent
	if err := json.Unmarshal(msg, &evt); err != nil {
		return models.Candle{}, "", false
	}

	// ONLY return candles that are closed
	if !evt.K.IsClosed {
		return models.Candle{}, "", false
	}

	candle := models.Candle{
		Open:      decimal.RequireFromString(evt.K.Open.String()),
		High:      decimal.RequireFromString(evt.K.High.String()),
		Low:       decimal.RequireFromString(evt.K.Low.String()),
		Close:     decimal.RequireFromString(evt.K.Close.String()),
		Volume:    decimal.RequireFromString(evt.K.Volume.String()),
		OpenTime:  evt.K.StartTime,
		CloseTime: evt.K.CloseTime,
	}
	key := evt.K.Symbol + "_" + evt.K.Interval
	return candle, key, true
}

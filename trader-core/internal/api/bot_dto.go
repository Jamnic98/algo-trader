package api

import (
	"time"
	"trader-core/internal/bot"
	"trader-core/internal/db/models"
)

type BotDTO struct {
	ID       string             `json:"id"`
	Mode     string             `json:"mode"`
	Base     string             `json:"base"`
	Quote    string             `json:"quote"`
	Interval string             `json:"interval"`
	Status   bot.BotStatus      `json:"status"`
	Started  *string            `json:"started,omitempty"`
	Lookback string             `json:"lookback"`
	Quantity string             `json:"quantity"`
	Candles  []models.CandleDTO `json:"candles,omitempty"`
}

func botToDTO(b *bot.Bot) BotDTO {
	var started *string
	if !b.Started.IsZero() {
		s := b.Started.Format(time.RFC3339)
		started = &s
	}

	candles := make([]models.CandleDTO, len(b.Candles))
	for i, c := range b.Candles {
		candles[i] = models.CandleToDTO(c)
	}

	return BotDTO{
		ID:       b.ID,
		Mode:     string(b.Mode),
		Interval: b.Interval.String(),
		Lookback: b.Lookback.String(),
		Started:  started,
		Status:   b.Status,
		Base:     b.Base,
		Quote:    b.Quote,
		Quantity: b.Quantity.String(),
		Candles:  candles,
	}
}

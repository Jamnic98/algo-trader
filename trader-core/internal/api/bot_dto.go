package api

import (
	"time"
	"trader-core/internal/bot"
	"trader-core/internal/db/models"
)

type BotDTO struct {
	ID           string             `json:"id"`
	Mode         string             `json:"mode"`
	Exchange     string             `json:"exchange"`
	StrategySlug string             `json:"strategy_slug"`
	StrategyName string             `json:"strategy_name"`
	AssetType    string             `json:"assetType"`
	Base         string             `json:"base"`
	Quote        string             `json:"quote"`
	Interval     string             `json:"interval"`
	MaxCandles   int                `json:"maxCandles"`
	Quantity     string             `json:"quantity"`
	Status       bot.BotStatus      `json:"status"`
	Started      *string            `json:"started,omitempty"`
	Candles      []models.CandleDTO `json:"candles,omitempty"`
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
		ID:           b.ID,
		Mode:         string(b.Mode),
		Exchange:     b.Exchange,
		StrategySlug: b.StrategySlug,
		StrategyName: b.StrategyName,
		AssetType:    b.AssetType,
		Base:         b.Base,
		Quote:        b.Quote,
		Interval:     b.Interval,
		MaxCandles:   b.MaxCandles,
		Quantity:     b.Quantity,
		Status:       b.Status,
		Started:      started,
		Candles:      candles,
	}
}

func configToDTO(cfg bot.BotConfig) BotDTO {
	return BotDTO{
		ID:           cfg.ID,
		Mode:         string(cfg.Mode),
		Exchange:     cfg.Exchange,
		AssetType:    cfg.AssetType,
		Base:         cfg.Base,
		Quote:        cfg.Quote,
		Interval:     cfg.Interval,
		MaxCandles:   cfg.MaxCandles,
		Quantity:     cfg.Quantity,
		StrategySlug: cfg.StrategySlug,
		Status:       "dead",
	}
}

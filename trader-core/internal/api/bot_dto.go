package api

import (
	"fmt"
	"time"
	"trader-core/internal/bot"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
)

type BotDTO struct {
	ID         string             `json:"id"`
	Mode       string             `json:"mode"`
	Exchange   string             `json:"exchange"`
	Strategy   *dto.StrategyDTO   `json:"strategy,omitempty"`
	StrategyID uint               `json:"strategy_id,string"`
	AssetType  string             `json:"assetType"`
	Base       string             `json:"base"`
	Quote      string             `json:"quote"`
	Interval   string             `json:"interval"`
	MaxCandles int                `json:"maxCandles"`
	Quantity   string             `json:"quantity"`
	Status     bot.BotStatus      `json:"status"`
	Started    *string            `json:"started,omitempty"`
	Candles    []models.CandleDTO `json:"candles,omitempty"`
}

func botToDTO(b *bot.Bot) (BotDTO, error) {
	var started *string
	if !b.Started.IsZero() {
		s := b.Started.Format(time.RFC3339)
		started = &s
	}

	var strategy models.Strategy
	if err := runtime.DB.First(&strategy, b.StrategyID).Error; err != nil {
		return BotDTO{}, fmt.Errorf("strategy %d not found: %w", b.StrategyID, err)
	}
	strategyDTO := dto.StrategyToDTO(&strategy)

	candles := make([]models.CandleDTO, len(b.Candles))
	for i, c := range b.Candles {
		candles[i] = models.CandleToDTO(c)
	}

	return BotDTO{
		ID:         b.ID,
		Mode:       string(b.Mode),
		Exchange:   b.Exchange,
		StrategyID: b.StrategyID,
		Strategy:   &strategyDTO,
		AssetType:  b.AssetType,
		Base:       b.Base,
		Quote:      b.Quote,
		Interval:   b.Interval,
		MaxCandles: b.MaxCandles,
		Quantity:   b.Quantity,
		Status:     b.Status,
		Started:    started,
		Candles:    candles,
	}, nil
}

func configToDTO(cfg bot.BotConfig) BotDTO {
	d := BotDTO{
		ID:         cfg.ID,
		Mode:       string(cfg.Mode),
		Exchange:   cfg.Exchange,
		StrategyID: cfg.StrategyID,
		AssetType:  cfg.AssetType,
		Base:       cfg.Base,
		Quote:      cfg.Quote,
		Interval:   cfg.Interval,
		MaxCandles: cfg.MaxCandles,
		Quantity:   cfg.Quantity,
		Status:     "dead",
	}

	if cfg.StrategyID != 0 {
		var strategy models.Strategy
		if err := runtime.DB.First(&strategy, cfg.StrategyID).Error; err == nil {
			s := dto.StrategyToDTO(&strategy)
			d.Strategy = &s
		}
	}

	return d
}

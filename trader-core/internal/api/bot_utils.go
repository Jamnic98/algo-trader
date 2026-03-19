package api

import (
	"fmt"
	"math"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
)

func getBot(id string) *bot.Bot {
	activeBotsMu.RLock()
	b, exists := activeBots[id]
	activeBotsMu.RUnlock()
	if !exists {
		return nil
	}
	return b
}

func fetchBotTrades(botID string, page, limit int) ([]dto.TradeDTO, int64, int, error) {
	var trades []models.Trade
	var total int64

	db.DB.Model(&models.Trade{}).Where("bot_id = ?", botID).Count(&total)
	if err := db.DB.Where("bot_id = ?", botID).Order("timestamp DESC").Offset((page - 1) * limit).Limit(limit).Find(&trades).Error; err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	dtos := make([]dto.TradeDTO, len(trades))
	for i, t := range trades {
		dtos[i] = dto.TradeToDTO(&t)
	}

	return dtos, total, totalPages, nil
}

func parseBotCreateRequest(data bot.CreateBotData) (bot.BotConfig, error) {
	if data.Mode == "" {
		data.Mode = bot.BotModePaper
	}
	if data.Exchange == "" {
		data.Exchange = "binance"
	}
	if data.Interval == "" {
		data.Interval = "1h"
	}
	if data.MaxCandles == 0 {
		data.MaxCandles = 200
	}
	if data.AssetType == "" {
		data.AssetType = "crypto"
	}

	// now validate
	switch data.Exchange {
	case "binance":
		switch data.AssetType {
		case "crypto":
			if data.Base == "" || data.Quote == "" {
				return bot.BotConfig{}, fmt.Errorf("crypto requires base and quote")
			}
			// TODO: Stocks Implementation
		// case "stocks":
		// 	if data.Base == "" {
		// 		return bot.BotConfig{}, fmt.Errorf("stocks requires base (ticker)")
		// 	}
		// 	data.Quote = ""
		default:
			return bot.BotConfig{}, fmt.Errorf("unknown asset type %q", data.AssetType)
		}
	default:
		return bot.BotConfig{}, fmt.Errorf("unknown exchange %q", data.Exchange)
	}

	if data.Strategy.Name == "" {
		return bot.BotConfig{}, fmt.Errorf("strategy name is required")
	}

	interval, err := engine.ParseInterval(data.Interval)
	if err != nil {
		return bot.BotConfig{}, fmt.Errorf("invalid interval %q: %w", data.Interval, err)
	}

	lookback := interval.Lookback(data.MaxCandles).String()

	return bot.BotConfig{
		Mode:           data.Mode,
		Exchange:       data.Exchange,
		AssetType:      data.AssetType,
		Base:           data.Base,
		Quote:          data.Quote,
		StrategyConfig: data.Strategy,
		Interval:       data.Interval,
		MaxCandles:     data.MaxCandles,
		Quantity:       data.Quantity,
		Lookback:       lookback,
	}, nil
}

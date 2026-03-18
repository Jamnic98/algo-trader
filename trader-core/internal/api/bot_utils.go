package api

import (
	"fmt"
	"math"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
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
	if data.Lookback == "" {
		data.Lookback = "200h"
	}
	if data.Quantity == "" {
		data.Quantity = "0.001"
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
		case "stocks":
			if data.Base == "" {
				return bot.BotConfig{}, fmt.Errorf("stocks requires base (ticker)")
			}
			data.Quote = ""
		default:
			return bot.BotConfig{}, fmt.Errorf("unknown asset type %q", data.AssetType)
		}
	default:
		return bot.BotConfig{}, fmt.Errorf("unknown exchange %q", data.Exchange)
	}

	if data.Strategy.Name == "" {
		return bot.BotConfig{}, fmt.Errorf("strategy name is required")
	}
	if _, err := engine.ParseInterval(data.Interval); err != nil {
		return bot.BotConfig{}, fmt.Errorf("invalid interval %q: %w", data.Interval, err)
	}
	if _, err := time.ParseDuration(data.Lookback); err != nil {
		return bot.BotConfig{}, fmt.Errorf("invalid lookback %q: %w", data.Lookback, err)
	}
	if _, err := decimal.NewFromString(data.Quantity); err != nil {
		return bot.BotConfig{}, fmt.Errorf("invalid quantity %q: %w", data.Quantity, err)
	}

	return bot.BotConfig{
		Mode:           data.Mode,
		Exchange:       data.Exchange,
		AssetType:      data.AssetType,
		Base:           data.Base,
		Quote:          data.Quote,
		StrategyConfig: data.Strategy,
		Interval:       data.Interval,
		Lookback:       data.Lookback,
		Quantity:       data.Quantity,
	}, nil
}

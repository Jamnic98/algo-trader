package api

import (
	"math"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
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

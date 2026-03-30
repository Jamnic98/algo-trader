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

func fetchBotFills(botID string, page, limit int) ([]dto.FIllDTO, int64, int, error) {
	var fills []models.Fill
	var total int64

	db.DB.Model(&models.Fill{}).Where("bot_id = ?", botID).Count(&total)
	if err := db.DB.Where("bot_id = ?", botID).Order("timestamp DESC").Offset((page - 1) * limit).Limit(limit).Find(&fills).Error; err != nil {
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	dtos := make([]dto.FIllDTO, len(fills))
	for i, t := range fills {
		dtos[i] = dto.FIllToDTO(&t)
	}

	return dtos, total, totalPages, nil
}

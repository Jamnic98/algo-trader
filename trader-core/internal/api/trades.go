package api

import (
	"net/http"
	"time"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegisterTradeRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getTradesHandler)
}

type TradeQuery struct {
	// Filters
	BotID    string `form:"botId"`
	Symbol   string `form:"symbol"`
	Side     string `form:"side"`
	Exchange string `form:"exchange"`
	Base     string `form:"base"`
	Quote    string `form:"quote"`
	DateFrom string `form:"dateFrom"` // RFC3339 e.g. 2024-01-01T00:00:00Z
	DateTo   string `form:"dateTo"`

	// Pagination
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type PaginatedTradesResponse struct {
	Trades     []dto.TradeDTO `json:"trades"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int64          `json:"total_pages"`
}

func getTradesHandler(c *gin.Context) {
	var q TradeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 200 {
		q.Limit = 20
	}

	query := db.DB.Model(&models.Trade{})

	// Apply filters
	if q.BotID != "" {
		query = query.Where("bot_id = ?", q.BotID)
	}
	if q.Symbol != "" {
		query = query.Where("symbol = ?", q.Symbol)
	}
	if q.Side != "" {
		query = query.Where("side = ?", q.Side)
	}
	if q.Exchange != "" {
		query = query.Where("exchange = ?", q.Exchange)
	}
	if q.Base != "" {
		query = query.Where("base = ?", q.Base)
	}
	if q.Quote != "" {
		query = query.Where("quote = ?", q.Quote)
	}
	if q.DateFrom != "" {
		if t, err := time.Parse(time.RFC3339, q.DateFrom); err == nil {
			query = query.Where("timestamp >= ?", t)
		}
	}
	if q.DateTo != "" {
		if t, err := time.Parse(time.RFC3339, q.DateTo); err == nil {
			query = query.Where("timestamp <= ?", t)
		}
	}

	// Count total before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch page
	var trades []models.Trade
	offset := (q.Page - 1) * q.Limit
	if err := query.
		Order("timestamp DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tradesDTO := make([]dto.TradeDTO, len(trades))
	for i, t := range trades {
		tradesDTO[i] = dto.TradeToDTO(&t)
	}

	pages := total / int64(q.Limit)
	if total%int64(q.Limit) != 0 {
		pages++
	}

	c.JSON(http.StatusOK, PaginatedTradesResponse{
		Trades:     tradesDTO,
		Total:      total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: pages,
	})
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/gin-gonic/gin"
)

func RegisterFillRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getFillsHandler)
	rg.GET("/stream", streamAllFillsHandler)
}

type FillQuery struct {
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

type PaginatedFillsResponse struct {
	Fills      []dto.FIllDTO `json:"fills"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int64         `json:"total_pages"`
}

func getFillsHandler(c *gin.Context) {
	var q FillQuery
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

	query := db.DB.Model(&models.Fill{})

	// Apply filters
	if q.BotID != "" {
		query = query.Where("bot_id = ?", q.BotID)
	}
	if q.Symbol != "" {
		query = query.Where("symbol LIKE ?", "%"+strings.ToUpper(q.Symbol)+"%")
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
	var fills []models.Fill
	offset := (q.Page - 1) * q.Limit
	if err := query.
		Order("timestamp DESC").
		Limit(q.Limit).
		Offset(offset).
		Find(&fills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fillsDTO := make([]dto.FIllDTO, len(fills))
	for i, t := range fills {
		fillsDTO[i] = dto.FIllToDTO(&t)
	}

	pages := total / int64(q.Limit)
	if total%int64(q.Limit) != 0 {
		pages++
	}

	c.JSON(http.StatusOK, PaginatedFillsResponse{
		Fills:      fillsDTO,
		Total:      total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: pages,
	})
}

func streamAllFillsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// subscribe to all active bots
	type sub struct {
		bot *bot.Bot
		ch  chan dto.FIllDTO
	}
	var subs []sub

	activeBotsMu.RLock()
	for _, b := range activeBots {
		subs = append(subs, sub{bot: b, ch: b.FillBroadcaster.Subscribe()})
	}
	activeBotsMu.RUnlock()

	defer func() {
		for _, s := range subs {
			s.bot.FillBroadcaster.Unsubscribe(s.ch)
		}
	}()

	// merge all channels into one
	merged := make(chan dto.FIllDTO)
	for _, s := range subs {
		go func(ch <-chan dto.FIllDTO) {
			for fill := range ch {
				merged <- fill
			}
		}(s.ch)
	}

	for {
		select {
		case fill := <-merged:
			data, _ := json.Marshal(fill)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

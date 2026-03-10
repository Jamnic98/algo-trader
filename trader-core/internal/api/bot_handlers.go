package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"trader-core/internal/bot"

	"github.com/gin-gonic/gin"
)

var (
	runtime      *bot.Runtime
	activeBotsMu sync.RWMutex
	activeBots   = make(map[string]*bot.Bot)
)

func InitBotAPI(rt *bot.Runtime) {
	runtime = rt
}

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("", getBotsHandler)
	rg.GET("/:id", getBotByIDHandler)
	rg.GET("/:id/trades", getBotTradesHandler)
	rg.GET("/:id/trades/stream", streamBotTradesHandler)

	rg.POST("", createBotHandler)
	rg.POST("/:id/start", startBotHandler)
	rg.POST("/:id/stop", stopBotHandler)
	rg.POST("/:id/attach", attachBotHandler)
	rg.POST("/:id/detach", detachBotHandler)
	rg.DELETE("/:id", deleteBotHandler)
}

func getBotByIDHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func getBotsHandler(c *gin.Context) {
	activeBotsMu.RLock()
	bots := make([]*bot.Bot, 0, len(activeBots))
	for _, b := range activeBots {
		bots = append(bots, b)
	}
	activeBotsMu.RUnlock()

	dtos := make([]BotDTO, len(bots))
	for i, b := range bots {
		dtos[i] = botToDTO(b)
	}

	c.JSON(http.StatusOK, gin.H{"bots": dtos})
}

func getBotTradesHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	dtos, total, totalPages, err := fetchBotTrades(c.Param("id"), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dtos,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func streamBotTradesHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	ch := b.TradeBroadcaster.Subscribe()
	defer b.TradeBroadcaster.Unsubscribe(ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for {
		select {
		case trade := <-ch:
			data, _ := json.Marshal(trade)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func createBotHandler(c *gin.Context) {
	var req struct {
		Symbol   string `json:"symbol"`
		Interval string `json:"interval"`
		Lookback string `json:"lookback"`
		Quantity string `json:"quantity"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := parseBotCreateRequest(req.Symbol, req.Interval, req.Lookback, req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := runtime.BotFactory.NewPaperBot(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBotsMu.Lock()
	activeBots[b.ID] = b
	activeBotsMu.Unlock()

	c.JSON(http.StatusCreated, gin.H{"bot": botToDTO(b)})
}

func startBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := b.Start(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func attachBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := runtime.AttachBot(b); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func detachBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := runtime.DetachBot(b); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func stopBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	b.Stop()
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func deleteBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if b.Status == bot.BotRunning {
		b.Stop()
	}
	if b.Status == bot.BotAttached {
		if err := runtime.DetachBot(b); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
	}

	activeBotsMu.Lock()
	delete(activeBots, b.ID)
	activeBotsMu.Unlock()

	c.Status(http.StatusNoContent)
}

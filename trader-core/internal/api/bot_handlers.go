package api

import (
	"encoding/json"
	"fmt"
	"log"
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

	cfgs, err := runtime.LoadBots()
	if err != nil {
		log.Printf("failed to load bots: %v", err)
		return
	}
	for _, cfg := range cfgs {
		b, err := runtime.RestoreBot(cfg)
		if err != nil {
			log.Printf("failed to restore bot %s: %v", cfg.ID, err)
			continue
		}
		activeBotsMu.Lock()
		activeBots[b.ID] = b
		activeBotsMu.Unlock()
	}
}

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("", getBotsHandler)
	rg.GET("/:id", getBotByIDHandler)
	rg.GET("/:id/trades", getBotTradesHandler)
	rg.GET("/:id/trades/stream", streamBotTradesHandler)
	rg.GET("/:id/logs/stream", streamBotLogsHandler)

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
	includeDeleted := c.Query("deleted") == "true"

	var cfgs []bot.BotConfig
	if includeDeleted {
		runtime.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&cfgs)
	} else {
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
		return
	}

	dtos := make([]BotDTO, len(cfgs))
	for i, cfg := range cfgs {
		dtos[i] = configToDTO(cfg)
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

func streamBotLogsHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	snapshot, ch := b.Logger.Subscribe()
	defer b.Logger.Unsubscribe(ch)

	// send catch-up buffer first
	for _, entry := range snapshot {
		data, _ := json.Marshal(entry)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}
	c.Writer.Flush()

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func createBotHandler(c *gin.Context) {
	var req struct {
		Base     string      `json:"base"`
		Quote    string      `json:"quote"`
		Interval string      `json:"interval"`
		Lookback string      `json:"lookback"`
		Quantity string      `json:"quantity"`
		Mode     bot.BotMode `json:"mode"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := parseBotCreateRequest(req.Base, req.Quote, req.Interval, req.Lookback, req.Quantity, req.Mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := runtime.CreateBot(cfg)
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

	if err := runtime.DeleteBot(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBotsMu.Lock()
	delete(activeBots, b.ID)
	activeBotsMu.Unlock()

	c.Status(http.StatusNoContent)
}

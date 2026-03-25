package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"trader-core/internal/bot"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
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
	rg.GET("/:id/positions", getBotPositionsHandler)
	rg.GET("/:id/candles/stream", streamBotCandlesHandler)
	rg.GET("/:id/ticks/stream", streamBotTicksHandler)

	rg.POST("", createBotHandler)
	rg.POST("/:id/start", startBotHandler)
	rg.POST("/:id/stop", stopBotHandler)
	rg.DELETE("/:id", deleteBotHandler)
}

func getBotByIDHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	d, err := botToDTO(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": d})
}

func getBotsHandler(c *gin.Context) {
	includeDeleted := c.Query("deleted") == "true"

	if includeDeleted {
		var cfgs []bot.BotConfig
		runtime.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&cfgs)
		dtos := make([]BotDTO, len(cfgs))
		for i, cfg := range cfgs {
			dtos[i] = configToDTO(cfg)
		}
		c.JSON(http.StatusOK, gin.H{"bots": dtos})
		return
	}

	activeBotsMu.RLock()
	bots := make([]*bot.Bot, 0, len(activeBots))
	for _, b := range activeBots {
		bots = append(bots, b)
	}
	activeBotsMu.RUnlock()

	dtos := make([]BotDTO, 0, len(bots))
	for _, b := range bots {
		d, err := botToDTO(b)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dtos = append(dtos, d)
	}
	c.JSON(http.StatusOK, gin.H{"bots": dtos})
}

func createBotHandler(c *gin.Context) {
	var req bot.CreateBotData
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity != "" {
		if _, err := decimal.NewFromString(req.Quantity); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid quantity %q: must be a valid number", req.Quantity)})
			return
		}
	}

	b, err := runtime.CreateBot(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activeBotsMu.Lock()
	activeBots[b.ID] = b
	activeBotsMu.Unlock()

	d, err := botToDTO(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"bot": d})
}

func startBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := runtime.AttachBot(b); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	d, err := botToDTO(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": d})
}

func stopBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}
	if err := runtime.DetachBot(b); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	d, err := botToDTO(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bot": d})
}

func deleteBotHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	if b.Status == bot.BotRunning {
		runtime.DetachBot(b)
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

func getBotTradesHandler(c *gin.Context) {
	botId := c.Param("id")
	if getBot(botId) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	dtos, total, totalPages, err := fetchBotTrades(botId, page, limit)
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

func getBotPositionsHandler(c *gin.Context) {
	botId := c.Param("id")
	if getBot(botId) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "positions not found"})
		return
	}

	positions, err := dto.CalcBotPositions(botId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"positions": positions})
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

	for _, entry := range snapshot {
		data, _ := json.Marshal(entry)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}

	fmt.Fprintf(c.Writer, "event: ready\ndata: {}\n\n")
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

func streamBotCandlesHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch := b.CandleBroadcaster.Subscribe()
	defer b.CandleBroadcaster.Unsubscribe(ch)

	for _, candle := range b.Candles {
		data, _ := json.Marshal(models.CandleToDTO(candle))
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	}

	fmt.Fprintf(c.Writer, "event: ready\ndata: {}\n\n")
	c.Writer.Flush()

	for {
		select {
		case candle := <-ch:
			data, _ := json.Marshal(models.CandleToDTO(candle))
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func streamBotTicksHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch := b.TickBroadcaster.Subscribe()
	defer b.TickBroadcaster.Unsubscribe(ch)

	fmt.Fprintf(c.Writer, "event: ready\ndata: {}\n\n")
	c.Writer.Flush()

	for {
		select {
		case candle := <-ch:
			data, _ := json.Marshal(models.CandleToDTO(candle))
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

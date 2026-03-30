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
	rg.GET("/:id/fills", getBotFillsHandler)
	rg.GET("/:id/fills/stream", streamBotFillsHandler)
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
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func getBotsHandler(c *gin.Context) {
	includeDeleted := c.Query("deleted") == "true"

	var cfgs []bot.BotConfig
	if includeDeleted {
		runtime.DB.Unscoped().Preload("Strategy").Where("deleted_at IS NOT NULL").Find(&cfgs)
		dtos := make([]BotDTO, len(cfgs))
		for i, cfg := range cfgs {
			dtos[i] = configToDTO(cfg)
		}
		c.JSON(http.StatusOK, gin.H{"bots": dtos})
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
	}
}

func getBotFillsHandler(c *gin.Context) {
	botId := c.Param("id")
	if getBot(botId) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	dtos, total, totalPages, err := fetchBotFills(botId, page, limit)
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

func streamBotFillsHandler(c *gin.Context) {
	b := getBot(c.Param("id"))
	if b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	ch := b.FillBroadcaster.Subscribe()
	defer b.FillBroadcaster.Unsubscribe(ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for {
		select {
		case fill := <-ch:
			data, _ := json.Marshal(fill)
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

	// ready event unblocks frontend — no log message, no duplicates
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

	// send existing candles as snapshot first
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

	// added ready event — was missing
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

	c.JSON(http.StatusCreated, gin.H{"bot": botToDTO(b)})
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
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
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
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
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

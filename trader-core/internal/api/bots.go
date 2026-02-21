package api

import (
	"net/http"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
)

type BotDTO struct {
	ID       string        `json:"id"`
	Symbol   string        `json:"symbol"`
	Interval string        `json:"interval"`
	Status   bot.BotStatus `json:"status"`
	Started  *string       `json:"started,omitempty"`
	Lookback string        `json:"lookback"`
}

func botToDTO(b *bot.Bot) BotDTO {
	var started *string
	if !b.Started.IsZero() {
		s := b.Started.Format(time.RFC3339)
		started = &s
	}

	return BotDTO{
		ID:       b.ID,
		Interval: b.Interval.String(),
		Lookback: b.Lookback.String(),
		Started:  started,
		Status:   b.Status,
		Symbol:   b.Symbol,
	}
}

var (
	runtime    *bot.Runtime
	activeBots = make(map[string]*bot.Bot)
)

func InitBotAPI(rt *bot.Runtime) {
	runtime = rt
}

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getBotsHandler)
	rg.GET("/:id/", getBotByIDHandler)
	rg.POST("/", createBotHandler)
	rg.POST("/:id/start/", startBotHandler)
	rg.POST("/:id/stop/", stopBotHandler)
	rg.POST("/:id/attach/", attachBotHandler)
	rg.POST("/:id/detach/", detachBotHandler)
	rg.DELETE("/:id/", deleteBotHandler)
}

func getBotsHandler(c *gin.Context) {
	bots := []BotDTO{}
	for _, b := range activeBots {
		bots = append(bots, botToDTO(b))
	}

	c.JSON(http.StatusOK, gin.H{"bots": bots})
}

func getBotByIDHandler(c *gin.Context) {
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func createBotHandler(c *gin.Context) {
	var req struct {
		Symbol   string `json:"symbol"`
		Interval string `json:"interval"`
		Lookback string `json:"lookback"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lookback, err := time.ParseDuration(req.Lookback)
	if err != nil || lookback <= 0 {
		lookback = 24 * time.Hour
	}

	interval, err := engine.ParseInterval(req.Interval)
	if err != nil {
		interval = engine.Interval1m
	}

	b, err := runtime.BotFactory.NewPaperBot(bot.BotConfig{
		Symbol:   req.Symbol,
		Interval: interval,
		Lookback: lookback,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBots[b.ID] = b
	c.JSON(http.StatusCreated, gin.H{"bot": botToDTO(b)})
}

func startBotHandler(c *gin.Context) {
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
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
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
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
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
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
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	b.Stop()
	c.JSON(http.StatusOK, gin.H{"bot": botToDTO(b)})
}

func deleteBotHandler(c *gin.Context) {
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
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

	delete(activeBots, id)
	c.Status(http.StatusNoContent)
}

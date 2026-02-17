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
	Started  string        `json:"started"`
	Lookback string        `json:"lookback"`
}

func botToDTO(b *bot.Bot) BotDTO {
	return BotDTO{
		ID:       b.ID,
		Interval: b.Interval.String(),
		Lookback: b.Lookback.String(),
		Started:  b.Started.Format(time.RFC3339),
		Status:   b.Status,
		Symbol:   b.Symbol,
	}
}

var (
	runtime      *bot.Runtime
	paperAccount *engine.PaperAccount
	activeBots   = make(map[string]*bot.Bot)
)

func InitBotAPI(rt *bot.Runtime, acc *engine.PaperAccount) {
	runtime = rt
	paperAccount = acc
}

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getBotsHandler)
	rg.GET("/:id/", getBotByIDHandler)
	rg.POST("/", createBotHandler)
	rg.POST("/:id/start/", startBotHandler)
	rg.POST("/:id/stop/", stopBotHandler)
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

	// Parse duration
	lookback, err := time.ParseDuration(req.Lookback)
	if err != nil || lookback <= 0 {
		lookback = 24 * time.Hour
	}

	interval, err := engine.ParseInterval(req.Interval)
	if err != nil {
		interval = engine.Interval1m
	}

	// Create bot via factory
	botFactory := bot.BotFactory{PaperAccount: paperAccount}
	b, err := botFactory.NewPaperBot(bot.BotConfig{
		Symbol:   req.Symbol,
		Interval: interval,
		Lookback: lookback,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Attach bot to runtime (dispatcher + market manager)
	if err := runtime.AttachBot(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeBots[b.ID] = b
	c.JSON(http.StatusCreated, gin.H{"bot": botToDTO(b)})
}

func deleteBotHandler(c *gin.Context) {
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

	b.Stop()
	delete(activeBots, id)

	c.Status(http.StatusNoContent)
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

	c.Status(http.StatusAccepted)
}

func stopBotHandler(c *gin.Context) {
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	b.Stop()

	c.Status(http.StatusNoContent)
}

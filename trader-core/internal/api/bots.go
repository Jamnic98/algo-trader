package api

import (
	"context"
	"net/http"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var activeBots = make(map[string]*bot.Bot)

func RegisterBotRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getBotsHandler)
	rg.GET("/:id", getBotByIDHandler)
	rg.POST("/", createBotHandler)
	rg.DELETE("/:id/", deleteBotHandler)
}

func getBotsHandler(c *gin.Context) {
	bots := []bot.Bot{}
	for _, b := range activeBots {
		bots = append(bots, *b)
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

	c.JSON(http.StatusOK, gin.H{"bot": b})
}

func createBotHandler(c *gin.Context) {
	var b bot.Bot
	if err := c.BindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ID if not provided
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	// Parse interval string from JSON
	dur, err := time.ParseDuration(b.Interval)
	if err != nil || dur <= 0 {
		dur = time.Minute // default to 1 minute
	}

	if _, exists := activeBots[b.ID]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "bot with this ID already exists"})
		return
	}

	b.Status = "running"
	b.Started = time.Now()

	// Context to cancel goroutine later
	ctx, cancel := context.WithCancel(context.Background())
	b.SetCancel(cancel)

	b.Engine = engine.NewPaperExecution(10000, 0.001)

	activeBots[b.ID] = &b

	// Launch strategy loop with parsed duration
	go bot.RunBotStrategy(ctx, &b)

	c.JSON(http.StatusCreated, gin.H{"bot": b})
}

func deleteBotHandler(c *gin.Context) {
	id := c.Param("id")
	b, exists := activeBots[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	// Stop the strategy goroutine using Stop()
	b.Stop()

	delete(activeBots, id)
	c.Status(http.StatusNoContent)
}

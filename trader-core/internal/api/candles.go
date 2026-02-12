package api

import "github.com/gin-gonic/gin"

type Candle struct {
    Symbol string  `json:"symbol"`
    Open   float64 `json:"open"`
    Close  float64 `json:"close"`
}

// Register routes for candles
func RegisterCandleRoutes(rg *gin.RouterGroup) {
    rg.GET("/", getCandlesHandler)
    rg.POST("/", createCandleHandler)
}

func getCandlesHandler(c *gin.Context) {
    // Fetch candles from DB
    c.JSON(200, gin.H{"candles": []string{}})
}

func createCandleHandler(c *gin.Context) {
    var candle Candle
    if err := c.BindJSON(&candle); err != nil {
        c.JSON(400, gin.H{"error": "invalid JSON"})
        return
    }

    // Store candle in DB
    c.JSON(201, gin.H{"status": "ok"})
}

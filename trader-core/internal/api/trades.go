package api

import (
	"net/http"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/gin-gonic/gin"
)

// Register routes for trades
func RegisterTradeRoutes(rg *gin.RouterGroup) {
    rg.GET("/", getTradesHandler)
    rg.POST("/", createTradeHandler)
}

func getTradesHandler(c *gin.Context) {
    var trades []models.Trade

    if err := db.DB.Find(&trades).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"trades": trades})
}

func createTradeHandler(c *gin.Context) {
    var input struct {
        symbol string 
        price  float64
    }

    if err := c.BindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
        return
    }

    trade := models.Trade{
        Symbol:    input.symbol,
        Price:     input.price,
        // Timestamp: time.Now(),
    }

    if err := db.DB.Create(&trade).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"trade":  trade})
}

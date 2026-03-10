package api

import (
	"net/http"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/gin-gonic/gin"
)

// Register routes for trades
func RegisterTradeRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getTradesHandler)
}

func getTradesHandler(c *gin.Context) {
	var trades []models.Trade

	if err := db.DB.Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tradesDTO := make([]dto.TradeDTO, len(trades))
	for i, t := range trades {
		tradesDTO[i] = dto.TradeToDTO(&t)
	}

	c.JSON(http.StatusOK, gin.H{"trades": tradesDTO})
}

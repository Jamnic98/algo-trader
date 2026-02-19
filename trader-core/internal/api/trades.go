package api

import (
	"net/http"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/gin-gonic/gin"
)

func tradeToDTO(t *models.Trade) models.TradeDTO {
	return models.TradeDTO{
		ID:        t.ID,
		BotID:     t.BotID,
		Symbol:    t.Symbol,
		Side:      t.Side,
		Price:     t.Price,
		Quantity:  t.Quantity,
		Fee:       t.Fee,
		FeeAsset:  t.FeeAsset,
		Exchange:  t.Exchange,
		Timestamp: t.Timestamp,
		CreatedAt: t.CreatedAt,
	}
}

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

	tradesDTO := make([]models.TradeDTO, len(trades))
	for i, t := range trades {
		tradesDTO[i] = tradeToDTO(&t)
	}

	c.JSON(http.StatusOK, gin.H{"trades": tradesDTO})
}

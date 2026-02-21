package api

import (
	"net/http"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	priceScale    = int64(1e8)
	quantityScale = int64(1e6)
	feeScale      = int64(1e8)
)

func tradeToDTO(t *models.Trade) models.TradeDTO {
	return models.TradeDTO{
		ID:        t.ID,
		BotID:     t.BotID,
		Symbol:    t.Symbol,
		Side:      t.Side,
		Price:     decimal.NewFromInt(t.PriceInt).Div(decimal.NewFromInt(priceScale)),
		Quantity:  decimal.NewFromInt(t.QuantityInt).Div(decimal.NewFromInt(quantityScale)),
		Fee:       decimal.NewFromInt(t.FeeInt).Div(decimal.NewFromInt(feeScale)),
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

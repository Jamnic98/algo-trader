package api

import (
	"net/http"
	"trader-core/internal/binance"

	"github.com/gin-gonic/gin"
)

func RegisterPriceRoutes(r *gin.RouterGroup) {
	r.GET("", getPriceHandler)
}

func getPriceHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	price, err := binance.FetchPrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"price": price})
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPriceRoutes(r *gin.RouterGroup) {
	r.GET("", getPriceHandler)
}

// getPriceHandler returns the latest price via the venue's broker:
// /api/price?symbol=BTCUSDT&exchange=binance (default) or symbol=BTC/USD&exchange=alpaca.
func getPriceHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	exchangeName := c.Query("exchange")
	if exchangeName == "" {
		exchangeName = "binance"
	}

	if runtime == nil || runtime.Brokers == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "exchange registry not initialised"})
		return
	}
	br, err := runtime.BrokerFor(exchangeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	price, err := br.Price(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"price": price.String(), "exchange": br.Name()})
}

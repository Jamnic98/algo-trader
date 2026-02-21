package api

import (
	"net/http"

	"trader-core/internal/bot"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
)

type AccountResponse struct {
	Balance   string            `json:"balance"`
	Positions map[string]string `json:"positions"`
}

var account engine.Account

func InitAccountAPI(rt *bot.Runtime) {
	account = rt.Account
}

func RegisterAccountRoutes(rg *gin.RouterGroup) {
	rg.GET("/", getAccountSnapshot)
}

func getAccountSnapshot(c *gin.Context) {
	accountSnapshot := account.Snapshot()

	posStr := make(map[string]string, len(accountSnapshot.Positions))
	for sym, qty := range accountSnapshot.Positions {
		posStr[sym] = qty.String()
	}

	c.JSON(http.StatusOK, AccountResponse{
		Balance:   accountSnapshot.Balance.String(),
		Positions: posStr,
	})
}

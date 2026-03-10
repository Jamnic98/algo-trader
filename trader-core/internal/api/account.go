package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	rg.GET("", getAccountSnapshot)
	rg.GET("/stream", streamAccount)
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

func streamAccount(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// send immediately on connect
	sendAccountSnapshot(c)

	for {
		select {
		case <-ticker.C:
			sendAccountSnapshot(c)
		case <-c.Request.Context().Done():
			return
		}
	}
}

func sendAccountSnapshot(c *gin.Context) {
	snapshot := account.Snapshot()
	posStr := make(map[string]string, len(snapshot.Positions))
	for sym, qty := range snapshot.Positions {
		posStr[sym] = qty.String()
	}
	data, _ := json.Marshal(AccountResponse{
		Balance:   snapshot.Balance.String(),
		Positions: posStr,
	})
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

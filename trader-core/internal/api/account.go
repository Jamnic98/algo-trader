package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AccountResponse struct {
	Balance   string            `json:"balance"`
	Positions map[string]string `json:"positions"`
}

var (
	accounts  *engine.AccountManager
	account   engine.Account // default paper account (back-compat for old routes)
	liveQuote = "USDT"
)

func InitAccountAPI(rt *bot.Runtime) {
	accounts = rt.Accounts
	account = rt.Account
}

func RegisterAccountRoutes(rg *gin.RouterGroup) {
	// back-compat: default account snapshot + stream
	rg.GET("", getAccountSnapshot)
	rg.GET("/stream", streamAccount)

	// account management
	rg.GET("/all", listAccountsHandler)
	rg.GET("/live", getLiveAccountHandler)
	rg.POST("/live/sync", syncLiveAccountHandler)
	rg.GET("/:name", getAccountHandler)
	rg.POST("/paper", createPaperAccountHandler)
	rg.POST("/paper/:name/reset", resetPaperAccountHandler)
	rg.POST("/paper/:name/topup", topUpPaperAccountHandler)
}

func getAccountSnapshot(c *gin.Context) {
	if account == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no default account"})
		return
	}
	snapshot := account.Snapshot()
	c.JSON(http.StatusOK, AccountResponse{
		Balance:   snapshot.Balance.String(),
		Positions: positionsToString(snapshot.Positions),
	})
}

func streamAccount(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

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
	if account == nil {
		return
	}
	snapshot := account.Snapshot()
	data, _ := json.Marshal(AccountResponse{
		Balance:   snapshot.Balance.String(),
		Positions: positionsToString(snapshot.Positions),
	})
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

func positionsToString(positions map[string]decimal.Decimal) map[string]string {
	out := make(map[string]string, len(positions))
	for sym, qty := range positions {
		out[sym] = qty.String()
	}
	return out
}

func listAccountsHandler(c *gin.Context) {
	if accounts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "account manager not initialised"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts.List()})
}

func getAccountHandler(c *gin.Context) {
	if accounts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "account manager not initialised"})
		return
	}
	info, err := accounts.Get(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

type createPaperRequest struct {
	Name    string `json:"name" binding:"required"`
	Balance string `json:"balance"`
	Fee     string `json:"fee"`
}

func createPaperAccountHandler(c *gin.Context) {
	var req createPaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Balance == "" {
		req.Balance = "10000"
	}
	if req.Fee == "" {
		req.Fee = "0.001"
	}
	if _, err := decimal.NewFromString(req.Balance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid balance"})
		return
	}
	if _, err := decimal.NewFromString(req.Fee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fee"})
		return
	}
	acc, err := accounts.CreatePaper(req.Name, req.Balance, req.Fee)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	snap := acc.Snapshot()
	c.JSON(http.StatusCreated, gin.H{
		"name": req.Name, "type": engine.AccountTypePaper,
		"balance": snap.Balance.String(), "fee": acc.Fee.String(),
		"positions": positionsToString(snap.Positions),
	})
}

type resetPaperRequest struct {
	Balance string `json:"balance"`
	Fee     string `json:"fee"`
}

func resetPaperAccountHandler(c *gin.Context) {
	var req resetPaperRequest
	_ = c.ShouldBindJSON(&req) // empty body = keep current balance/fee? no: reset to defaults
	name := c.Param("name")

	info, err := accounts.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	balance := req.Balance
	if balance == "" {
		balance = "10000"
	}
	fee := req.Fee
	if fee == "" {
		fee = info.Fee
	}
	if err := accounts.ResetPaper(name, balance, fee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	info, _ = accounts.Get(name)
	c.JSON(http.StatusOK, info)
}

type topUpRequest struct {
	Amount string `json:"amount" binding:"required"`
}

func topUpPaperAccountHandler(c *gin.Context) {
	var req topUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be a positive decimal"})
		return
	}
	name := c.Param("name")
	if err := accounts.TopUp(name, amount); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	info, _ := accounts.Get(name)
	c.JSON(http.StatusOK, info)
}

// getLiveAccountHandler returns the live ledger merged with fresh Binance balances.
func getLiveAccountHandler(c *gin.Context) {
	if accounts == nil || accounts.Live() == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no live account configured"})
		return
	}
	info, err := accounts.Get("live")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// enrich with real balances from Binance (best effort)
	balances, berr := binance.FetchAccount()
	if berr == nil {
		enriched := map[string]string{}
		for asset, amt := range balances {
			enriched[asset] = amt.String()
		}
		info.Positions["_exchange_balances"] = "" // placeholder removed below
		delete(info.Positions, "_exchange_balances")
		c.JSON(http.StatusOK, gin.H{"account": info, "exchange_balances": enriched})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": info, "exchange_balances": nil, "exchange_error": berr.Error()})
}

// syncLiveAccountHandler pulls real Binance balances into the live ledger.
func syncLiveAccountHandler(c *gin.Context) {
	if accounts == nil || accounts.Live() == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no live account configured"})
		return
	}
	balances, err := binance.FetchAccount()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "binance account fetch failed: " + err.Error()})
		return
	}
	accounts.SyncLive(balances, liveQuote)
	info, _ := accounts.Get("live")
	c.JSON(http.StatusOK, gin.H{"account": info})
}

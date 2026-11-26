package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AccountResponse struct {
	Balance   string            `json:"balance"`
	Positions map[string]string `json:"positions"`
}

var (
	accounts *engine.AccountManager
	account  engine.Account // default paper account (back-compat for old routes)
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

// getLiveAccountHandler returns the live ledger for ?exchange= (default
// binance), enriched with fresh balances from that venue (best effort).
func getLiveAccountHandler(c *gin.Context) {
	if accounts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "account manager not initialised"})
		return
	}
	exchangeName := c.DefaultQuery("exchange", "binance")
	info, err := accounts.Get("live:" + exchangeName)
	if err != nil {
		// no ledger yet: create an empty one if the venue is registered
		// (enrichment below fills it with fresh balances); unknown venue -> 404
		if runtime == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if _, berr := runtime.BrokerFor(exchangeName); berr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		acc := accounts.Live(exchangeName)
		snap := acc.Snapshot()
		info = engine.AccountInfo{
			Name: "live:" + exchangeName, Type: engine.AccountTypeLive,
			Balance: snap.Balance.String(), Fee: acc.Fee.String(),
			Positions: positionsToString(snap.Positions),
		}
	}

	// enrich with real balances from that venue (best effort)
	if runtime != nil {
		if br, berr := runtime.BrokerFor(exchangeName); berr == nil {
			if balances, ferr := br.Balances(); ferr == nil {
				enriched := map[string]string{}
				for asset, amt := range balances {
					enriched[asset] = amt.String()
				}
				c.JSON(http.StatusOK, gin.H{"account": info, "exchange": exchangeName, "exchange_balances": enriched})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{"account": info, "exchange": exchangeName, "exchange_balances": nil, "exchange_error": ferr.Error()})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"account": info, "exchange": exchangeName, "exchange_balances": nil})
}

// syncLiveAccountHandler pulls real venue balances into the per-exchange live
// ledgers. ?exchange=alpaca syncs only that venue; no param syncs all
// registered brokers.
func syncLiveAccountHandler(c *gin.Context) {
	if accounts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "account manager not initialised"})
		return
	}
	if runtime == nil || len(runtime.Brokers) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no brokers registered"})
		return
	}

	exchangeName := c.Query("exchange")
	syncOne := func(name string) (engine.AccountInfo, error) {
		br, err := runtime.BrokerFor(name)
		if err != nil {
			return engine.AccountInfo{}, err
		}
		balances, err := br.Balances()
		if err != nil {
			return engine.AccountInfo{}, fmt.Errorf("%s account fetch failed: %w", name, err)
		}
		accounts.SyncLive(name, balances, exchange.QuoteAssetFor(name))
		return accounts.Get("live:" + name)
	}

	if exchangeName != "" {
		info, err := syncOne(exchangeName)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"account": info})
		return
	}

	synced := map[string]engine.AccountInfo{}
	errs := map[string]string{}
	for name := range runtime.Brokers {
		if info, err := syncOne(name); err != nil {
			errs[name] = err.Error()
		} else {
			synced[name] = info
		}
	}
	c.JSON(http.StatusOK, gin.H{"accounts": synced, "errors": errs})
}

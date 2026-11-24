package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
)

// BacktestRequest is forwarded to the Python backtester service.
type BacktestRequest struct {
	StrategySlug   string `json:"strategy_slug" binding:"required"`
	Exchange       string `json:"exchange"` // venue for market data, default "binance"
	Symbol         string `json:"symbol" binding:"required"` // venue-native symbol (BTCUSDT, BTC/USD, AAPL)
	Interval       string `json:"interval" binding:"required"`
	Start          string `json:"start" binding:"required"` // RFC3339, YYYY-MM-DD or unix ms
	End            string `json:"end" binding:"required"`
	InitialBalance string `json:"initial_balance"` // decimal string, default 10000
	FeeRate        string `json:"fee_rate"`        // decimal string, default 0.001
	Quantity       string `json:"quantity"`        // optional fixed qty when the strategy doesn't dictate one
}

func RegisterBacktestRoutes(rg *gin.RouterGroup, backtesterURL string) {
	rg.POST("", func(c *gin.Context) { runBacktestHandler(c, backtesterURL) })
}

// strategyLookup is a var so tests can stub the DB lookup.
var strategyLookup = func(slug string) error {
	var strategy models.Strategy
	return db.DB.First(&strategy, "slug = ?", slug).Error
}

func runBacktestHandler(c *gin.Context, backtesterURL string) {
	var req BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := engine.ParseInterval(req.Interval); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := strategyLookup(req.StrategySlug); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}

	startMs, err := parseTimeMs(req.Start)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start: " + err.Error()})
		return
	}
	endMs, err := parseTimeMs(req.End)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end: " + err.Error()})
		return
	}
	if endMs <= startMs {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be after start"})
		return
	}

	if req.InitialBalance == "" {
		req.InitialBalance = "10000"
	}
	if req.FeeRate == "" {
		req.FeeRate = "0.001"
	}
	if req.Exchange == "" {
		req.Exchange = "binance"
	}

	payload := map[string]any{
		"strategy_slug":   req.StrategySlug,
		"exchange":        req.Exchange,
		"symbol":          req.Symbol,
		"interval":        req.Interval,
		"start":           startMs,
		"end":             endMs,
		"initial_balance": req.InitialBalance,
		"fee_rate":        req.FeeRate,
	}
	if req.Quantity != "" {
		payload["quantity"] = req.Quantity
	}

	status, body, err := postBacktest(c.Request.Context(), backtesterURL, payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "backtester unavailable: " + err.Error()})
		return
	}

	c.Data(status, "application/json", body)
}

// postBacktest forwards a backtest job to the Python backtester service.
func postBacktest(ctx context.Context, baseURL string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		strings.TrimRight(baseURL, "/")+"/backtest",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, respBody, nil
}

// parseTimeMs accepts RFC3339, YYYY-MM-DD (UTC) or unix milliseconds.
func parseTimeMs(s string) (int64, error) {
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		return ms, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UnixMilli(), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC().UnixMilli(), nil
	}
	return 0, fmt.Errorf("expected RFC3339, YYYY-MM-DD or unix ms, got %q", s)
}

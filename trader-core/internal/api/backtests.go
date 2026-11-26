package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	StrategySlug   string   `json:"strategy_slug" binding:"required"`
	Exchange       string   `json:"exchange"`   // venue for market data, default "binance"
	Symbol         string   `json:"symbol"`     // single venue-native symbol (BTCUSDT, BTC/USD, AAPL)
	Symbols        []string `json:"symbols"`    // multi-symbol portfolio run (overrides symbol)
	Interval       string   `json:"interval" binding:"required"`
	Start          string   `json:"start" binding:"required"` // RFC3339, YYYY-MM-DD or unix ms
	End            string   `json:"end" binding:"required"`
	InitialBalance string   `json:"initial_balance"` // decimal string, default 10000
	FeeRate        string   `json:"fee_rate"`        // decimal string, default 0.001
	SlippageBps    string   `json:"slippage_bps"`    // adverse slippage on fills in bps, default "0"
	Quantity       string   `json:"quantity"`        // optional fixed qty when the strategy doesn't dictate one
}

// SweepRequest runs the same backtest over a cartesian product of param
// overrides (handled by the backtester's /backtest/sweep).
type SweepRequest struct {
	BacktestRequest
	ParamGrid     map[string][]any `json:"param_grid" binding:"required"`
	MaxCombos     int              `json:"max_combos"` // safety cap, default 200
	IncludeTrades bool             `json:"include_trades"`
	SortBy        string           `json:"sort_by"` // total_return_pct|sharpe|profit_factor|max_drawdown_pct|num_trades
}

func RegisterBacktestRoutes(rg *gin.RouterGroup, backtesterURL string) {
	rg.POST("", func(c *gin.Context) { runBacktestHandler(c, backtesterURL) })
	rg.POST("/sweep", func(c *gin.Context) { runSweepHandler(c, backtesterURL) })
}

// strategyLookup is a var so tests can stub the DB lookup.
var strategyLookup = func(slug string) error {
	var strategy models.Strategy
	return db.DB.First(&strategy, "slug = ?", slug).Error
}

var errStrategyNotFound = errors.New("strategy not found")

// prepareBacktest validates the shared request fields and builds the payload
// forwarded to the Python backtester.
func prepareBacktest(req *BacktestRequest) (map[string]any, error) {
	if req.Symbol == "" && len(req.Symbols) == 0 {
		return nil, fmt.Errorf("symbol or symbols is required")
	}
	if _, err := engine.ParseInterval(req.Interval); err != nil {
		return nil, err
	}
	if err := strategyLookup(req.StrategySlug); err != nil {
		return nil, errStrategyNotFound
	}

	startMs, err := parseTimeMs(req.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start: %w", err)
	}
	endMs, err := parseTimeMs(req.End)
	if err != nil {
		return nil, fmt.Errorf("invalid end: %w", err)
	}
	if endMs <= startMs {
		return nil, fmt.Errorf("end must be after start")
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
	if req.SlippageBps == "" {
		req.SlippageBps = "0"
	}

	payload := map[string]any{
		"strategy_slug":   req.StrategySlug,
		"exchange":        req.Exchange,
		"interval":        req.Interval,
		"start":           startMs,
		"end":             endMs,
		"initial_balance": req.InitialBalance,
		"fee_rate":        req.FeeRate,
		"slippage_bps":    req.SlippageBps,
	}
	if req.Symbol != "" {
		payload["symbol"] = req.Symbol
	}
	if len(req.Symbols) > 0 {
		payload["symbols"] = req.Symbols
	}
	if req.Quantity != "" {
		payload["quantity"] = req.Quantity
	}
	return payload, nil
}

func writeBacktestError(c *gin.Context, err error) {
	if errors.Is(err, errStrategyNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func runBacktestHandler(c *gin.Context, backtesterURL string) {
	var req BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload, err := prepareBacktest(&req)
	if err != nil {
		writeBacktestError(c, err)
		return
	}
	forwardBacktest(c, backtesterURL, "/backtest", payload)
}

func runSweepHandler(c *gin.Context, backtesterURL string) {
	var req SweepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload, err := prepareBacktest(&req.BacktestRequest)
	if err != nil {
		writeBacktestError(c, err)
		return
	}
	payload["param_grid"] = req.ParamGrid
	if req.MaxCombos > 0 {
		payload["max_combos"] = req.MaxCombos
	}
	if req.IncludeTrades {
		payload["include_trades"] = true
	}
	if req.SortBy != "" {
		payload["sort_by"] = req.SortBy
	}
	forwardBacktest(c, backtesterURL, "/backtest/sweep", payload)
}

func forwardBacktest(c *gin.Context, backtesterURL, path string, payload any) {
	status, body, err := postBacktest(c.Request.Context(), backtesterURL, path, payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "backtester unavailable: " + err.Error()})
		return
	}
	c.Data(status, "application/json", body)
}

// postBacktest forwards a backtest job to the Python backtester service.
func postBacktest(ctx context.Context, baseURL, path string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		strings.TrimRight(baseURL, "/")+path,
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

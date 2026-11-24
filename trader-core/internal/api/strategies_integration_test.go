package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/strategies"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// TestCreateStrategyThenResolveIntegration is the UI → DB → backtester loop:
// it creates a composite strategy through the real HTTP handler against a
// real Postgres, reads the row back with the same query the Python backtester
// uses, and resolves it the same way the bot runtime does.
//
// Runs only when TEST_DB_DSN is set (CI provides a postgres service).
func TestCreateStrategyThenResolveIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping integration test")
	}

	db.ConnectPostgres(dsn)
	db.Migrate(&models.Strategy{})
	db.SeedStrategies(db.DB) // primitives the composite references
	db.DB.Exec("DELETE FROM strategies WHERE slug = ?", "itest_combo") // idempotent

	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterStrategyRoutes(r.Group("/api/strategies"))

	body := `{
		"slug": "itest_combo",
		"display_name": "ITest Combo",
		"threshold": 0.6,
		"nodes": [
			{"strategy_slug": "simple_dca", "weight": 1},
			{"strategy_slug": "simple", "weight": 1}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/strategies", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create strategy: status = %d, body = %s", w.Code, w.Body.String())
	}

	// Read the row back exactly the way the Python backtester's
	// PostgresStrategyStore does (same columns, same table).
	var (
		slug         string
		displayName  string
		isComposite  bool
		defaultCfg   []byte
		nodesJSON    []byte
	)
	err := db.DB.Raw(
		"SELECT slug, display_name, is_composite, default_config, nodes FROM strategies WHERE slug = ?",
		"itest_combo",
	).Row().Scan(&slug, &displayName, &isComposite, &defaultCfg, &nodesJSON)
	if err != nil {
		t.Fatalf("reading back row the backtester way: %v", err)
	}
	if !isComposite {
		t.Fatal("is_composite should be true")
	}
	// parse (jsonb re-serializes with different whitespace, so don't string-match)
	var cfg strategies.CompositeConfig
	if err := json.Unmarshal(defaultCfg, &cfg); err != nil {
		t.Fatalf("default_config not valid composite config: %v (%s)", err, defaultCfg)
	}
	if len(cfg.Nodes) != 2 || cfg.Nodes[0].StrategySlug != "simple_dca" || cfg.Threshold != 0.6 {
		t.Fatalf("default_config = %s", defaultCfg)
	}
	var legacyNodes []models.StrategyNode
	if err := json.Unmarshal(nodesJSON, &legacyNodes); err != nil || len(legacyNodes) != 2 {
		t.Fatalf("nodes column = %s (err %v)", nodesJSON, err)
	}

	// Resolve the row the same way the bot runtime does.
	var row models.Strategy
	if err := db.DB.First(&row, "slug = ?", "itest_combo").Error; err != nil {
		t.Fatalf("load row: %v", err)
	}
	strat, err := bot.ResolveStrategy(db.DB, row)
	if err != nil {
		t.Fatalf("ResolveStrategy: %v", err)
	}

	// simple_dca always buys; simple holds on the first candle, buys on a down
	// candle. Composite threshold 0.6: first candle 1/2 votes = HOLD, second = BUY.
	fee := decimal.NewFromFloat(0.001)
	c := func(close int64) strategies.CandleContext {
		p := decimal.NewFromInt(close)
		return strategies.CandleContext{
			Candle:  models.Candle{Close: p, Open: p, High: p, Low: p},
			FeeRate: fee,
		}
	}
	if d := strat.OnCandle(c(100)); d.Signal != strategies.Hold {
		t.Fatalf("candle 1: %v, want HOLD", d.Signal)
	}
	if d := strat.OnCandle(c(90)); d.Signal != strategies.Buy {
		t.Fatalf("candle 2: %v, want BUY", d.Signal)
	}
}

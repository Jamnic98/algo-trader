package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"trader-core/internal/bot"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
)

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origRuntime := runtime
	origBots := activeBots
	runtime = &bot.Runtime{Accounts: engine.NewAccountManager(nil, engine.NewLiveAccount("0", "0.001"))}
	activeBots = map[string]*bot.Bot{
		"b1": {BotConfig: bot.BotConfig{ID: "b1"}, Status: bot.BotRunning},
		"b2": {BotConfig: bot.BotConfig{ID: "b2"}, Status: bot.BotCreated},
	}
	t.Cleanup(func() { runtime = origRuntime; activeBots = origBots })

	r := gin.New()
	RegisterMetricsRoutes(r.Group("/api/metrics"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["bots_total"] != float64(2) || body["bots_running"] != float64(1) {
		t.Fatalf("bots = %v/%v, want 2/1", body["bots_total"], body["bots_running"])
	}
	// no DB in unit tests -> fills_today reported as -1
	if body["fills_today"] != float64(-1) {
		t.Fatalf("fills_today = %v, want -1 without DB", body["fills_today"])
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %v", body["status"])
	}
}

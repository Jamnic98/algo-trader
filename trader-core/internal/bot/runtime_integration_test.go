package bot

import (
	"os"
	"testing"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
)

// TestRestoreBotResolvesStrategyBySlug is a regression test for the
// `First(&strategyModel, cfg.StrategySlug)` bug, which produced
// `WHERE simple` (slug treated as a column) instead of `WHERE slug = 'simple'`.
//
// Runs only when TEST_DB_DSN is set (CI provides a postgres service).
func TestRestoreBotResolvesStrategyBySlug(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping integration test")
	}

	// dedicated database per package so parallel test binaries don't race on AutoMigrate
	if err := db.ConnectTestDatabase(dsn, "bot_itest"); err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	db.Migrate(&models.Strategy{}, &BotConfig{})
	db.SeedStrategies(db.DB)

	acc := engine.NewPaperAccount("10000", "0.001")
	accounts := engine.NewAccountManager(db.DB, engine.NewLiveAccount("0", "0.001"))
	if _, err := accounts.Paper(engine.DefaultPaperName); err != nil {
		t.Fatalf("default paper account: %v", err)
	}
	rt := &Runtime{
		DB:       db.DB,
		Account:  acc,
		Accounts: accounts,
		BotFactory: &BotFactory{
			Account: acc,
			Engine:  func() engine.ExecutionEngine { return engine.NewPaperExecutor(acc) },
		},
	}

	b, err := rt.RestoreBot(BotConfig{ID: "itest-bot", StrategySlug: "simple"})
	if err != nil {
		t.Fatalf("RestoreBot(simple): %v", err)
	}
	if b.ActiveStrategy == nil {
		t.Fatal("strategy was not resolved")
	}

	// unknown slug must return a clean "not found" error, not a SQL error
	if _, err := rt.RestoreBot(BotConfig{ID: "itest-bot-2", StrategySlug: "does_not_exist"}); err == nil {
		t.Fatal("expected error for unknown strategy slug")
	}
}

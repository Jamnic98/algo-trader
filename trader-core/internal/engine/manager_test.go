package engine

import (
	"os"
	"testing"

	"trader-core/internal/db"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

func newTestManager() *AccountManager {
	live := NewLiveAccount("0", "0.001")
	return NewAccountManager(nil, live) // nil db = in-memory
}

func TestManagerCreateListGetPaper(t *testing.T) {
	m := newTestManager()

	acc, err := m.CreatePaper("alpha", "5000", "0.002")
	if err != nil {
		t.Fatalf("CreatePaper: %v", err)
	}
	if !acc.Snapshot().Balance.Equal(decimal.NewFromInt(5000)) {
		t.Fatalf("balance = %s", acc.Snapshot().Balance)
	}
	if !acc.Fee.Equal(decimal.NewFromFloat(0.002)) {
		t.Fatalf("fee = %s", acc.Fee)
	}

	// duplicate rejected
	if _, err := m.CreatePaper("alpha", "1", "0.001"); err == nil {
		t.Fatal("expected duplicate error")
	}

	// default account auto-created on first use
	def, err := m.Paper("")
	if err != nil {
		t.Fatalf("Paper(default): %v", err)
	}
	if def.Snapshot().Balance.String() != "10000" {
		t.Fatalf("default balance = %s", def.Snapshot().Balance)
	}

	// unknown non-default account errors
	if _, err := m.Paper("nope"); err == nil {
		t.Fatal("expected not found error")
	}

	infos := m.List()
	if len(infos) != 2 { // alpha + default + live? live included -> 3
		t.Logf("accounts: %+v", infos)
	}
	names := map[string]bool{}
	for _, i := range infos {
		names[i.Name] = true
	}
	if !names["alpha"] || !names["default"] || !names["live"] {
		t.Fatalf("list = %+v", infos)
	}

	info, err := m.Get("alpha")
	if err != nil || info.Type != AccountTypePaper || info.Balance != "5000" {
		t.Fatalf("Get(alpha) = %+v, %v", info, err)
	}
}

func TestManagerTopUpAndReset(t *testing.T) {
	m := newTestManager()
	m.CreatePaper("beta", "100", "0.001")

	if err := m.TopUp("beta", decimal.NewFromInt(50)); err != nil {
		t.Fatalf("TopUp: %v", err)
	}
	info, _ := m.Get("beta")
	if info.Balance != "150" {
		t.Fatalf("balance after topup = %s", info.Balance)
	}

	if err := m.TopUp("missing", decimal.NewFromInt(1)); err == nil {
		t.Fatal("expected error topping up unknown account")
	}

	// simulate a position, then reset
	acc, _ := m.Paper("beta")
	acc.Positions["BTCUSDT"] = decimal.NewFromInt(1)
	if err := m.ResetPaper("beta", "1000", "0.001"); err != nil {
		t.Fatalf("ResetPaper: %v", err)
	}
	info, _ = m.Get("beta")
	if info.Balance != "1000" || len(info.Positions) != 0 {
		t.Fatalf("after reset = %+v", info)
	}
}

func TestManagerSyncLive(t *testing.T) {
	m := newTestManager()
	m.SyncLive(map[string]decimal.Decimal{
		"USDT": decimal.NewFromFloat(2500.5),
		"BTC":  decimal.NewFromFloat(0.25),
		"XRP":  decimal.Zero,
	}, "USDT")

	info, err := m.Get("live")
	if err != nil {
		t.Fatalf("Get(live): %v", err)
	}
	if info.Balance != "2500.5" {
		t.Fatalf("live balance = %s", info.Balance)
	}
	if info.Positions["BTC"] != "0.25" {
		t.Fatalf("positions = %v", info.Positions)
	}
	if _, ok := info.Positions["XRP"]; ok {
		t.Fatal("zero-balance asset should not be a position")
	}
	if _, ok := info.Positions["USDT"]; ok {
		t.Fatal("quote asset should not be a position")
	}
}

// TestManagerPersistsPaperAcrossReload runs only with TEST_DB_DSN (CI postgres).
func TestManagerPersistsPaperAcrossReload(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}
	if err := db.ConnectTestDatabase(dsn, "engine_itest"); err != nil {
		t.Fatalf("connect: %v", err)
	}
	db.Migrate(&models.Account{})
	db.DB.Exec("DELETE FROM accounts WHERE name LIKE 'mgr_%'")

	m := NewAccountManager(db.DB, NewLiveAccount("0", "0.001"))
	acc, err := m.CreatePaper("mgr_test", "777", "0.001")
	if err != nil {
		t.Fatalf("CreatePaper: %v", err)
	}
	acc.AddBalance(decimal.NewFromInt(23)) // should persist via OnChange

	// simulate restart: new manager from same DB
	m2 := NewAccountManager(db.DB, NewLiveAccount("0", "0.001"))
	acc2, err := m2.Paper("mgr_test")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := acc2.Snapshot().Balance.String(); got != "800" {
		t.Fatalf("balance after reload = %s, want 800", got)
	}
}


// TestManagerPersistsPositionsAcrossReload: open paper positions must survive
// a restart, otherwise a bot holding a position can't sell after reboot.
func TestManagerPersistsPositionsAcrossReload(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}
	if err := db.ConnectTestDatabase(dsn, "engine_itest"); err != nil {
		t.Fatalf("connect: %v", err)
	}
	db.Migrate(&models.Account{})
	db.DB.Exec("DELETE FROM accounts WHERE name LIKE 'mgr_pos_%'")

	m := NewAccountManager(db.DB, NewLiveAccount("0", "0.001"))
	acc, err := m.CreatePaper("mgr_pos_test", "10000", "0.001")
	if err != nil {
		t.Fatalf("CreatePaper: %v", err)
	}
	pe := NewPaperExecutor(acc)
	if _, err := pe.ExecuteFill(Order{BotID: "b", Symbol: "BTCUSDT", Signal: "BUY", Price: decimal.NewFromInt(100), Qty: decimal.NewFromInt(2)}); err != nil {
		t.Fatalf("buy: %v", err)
	}

	// simulate restart
	m2 := NewAccountManager(db.DB, NewLiveAccount("0", "0.001"))
	acc2, err := m2.Paper("mgr_pos_test")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	snap := acc2.Snapshot()
	if !snap.Positions["BTCUSDT"].Equal(decimal.NewFromInt(2)) {
		t.Fatalf("position after reload = %v, want 2 BTCUSDT", snap.Positions)
	}
	// and the restored account can actually sell
	if _, err := NewPaperExecutor(acc2).ExecuteFill(Order{BotID: "b", Symbol: "BTCUSDT", Signal: "SELL", Price: decimal.NewFromInt(110), Qty: decimal.NewFromInt(2)}); err != nil {
		t.Fatalf("sell after reload: %v", err)
	}
}

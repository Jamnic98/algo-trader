package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	AccountTypePaper = "paper"
	AccountTypeLive  = "live"
	DefaultPaperName = "default"
)

// AccountInfo is the API-facing view of an account.
type AccountInfo struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Balance   string            `json:"balance"`
	Fee       string            `json:"fee"`
	Positions map[string]string `json:"positions"`
}

// AccountManager owns all trading accounts: any number of paper accounts
// (persisted to the DB so they survive restarts) plus one live ledger per
// exchange ("live:binance", "live:alpaca", ...), each mirroring that venue's
// real balances.
type AccountManager struct {
	mu    sync.RWMutex
	db    *gorm.DB // nil = in-memory only (tests)
	paper map[string]*PaperAccount
	live  map[string]*LiveAccount // exchange name (lowercase) -> ledger
}

// DefaultLiveExchange is the venue used for the back-compat "live" name.
const DefaultLiveExchange = "binance"

func NewAccountManager(db *gorm.DB, live *LiveAccount) *AccountManager {
	m := &AccountManager{db: db, paper: map[string]*PaperAccount{}, live: map[string]*LiveAccount{}}
	if live != nil {
		m.live[DefaultLiveExchange] = live
	}
	if db != nil {
		var rows []models.Account
		if err := db.Where("type = ?", AccountTypePaper).Find(&rows).Error; err == nil {
			for _, row := range rows {
				acc := NewPaperAccount(row.Balance, row.Fee)
				acc.RestorePositions(parsePositions(row.Positions), nil)
				m.attachPaper(row.Name, acc)
			}
		}
	}
	return m
}

// parsePositions decodes the persisted positions JSON (symbol -> qty string).
func parsePositions(raw datatypes.JSON) map[string]decimal.Decimal {
	out := map[string]decimal.Decimal{}
	if len(raw) == 0 {
		return out
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return out
	}
	for sym, q := range m {
		if d, err := decimal.NewFromString(q); err == nil {
			out[sym] = d
		}
	}
	return out
}

func (m *AccountManager) attachPaper(name string, acc *PaperAccount) {
	acc.OnChange = func() { m.persistPaper(name, acc) }
	m.paper[name] = acc
}

func (m *AccountManager) persistPaper(name string, acc *PaperAccount) {
	if m.db == nil {
		return
	}
	snap := acc.Snapshot()
	posJSON, _ := json.Marshal(positionsToString(snap.Positions))
	m.db.Model(&models.Account{}).Where("name = ?", name).Updates(map[string]any{
		"balance":   snap.Balance.String(),
		"fee":       acc.Fee.String(),
		"positions": datatypes.JSON(posJSON),
	})
}

// CreatePaper creates a new paper account and persists it. Fails if the name exists.
func (m *AccountManager) CreatePaper(name, balance, fee string) (*PaperAccount, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.paper[name]; exists {
		return nil, fmt.Errorf("account %q already exists", name)
	}
	acc := NewPaperAccount(balance, fee)
	if m.db != nil {
		row := models.Account{Name: name, Type: AccountTypePaper, Balance: acc.Snapshot().Balance.String(), Fee: acc.Fee.String()}
		if err := m.db.Create(&row).Error; err != nil {
			return nil, err
		}
	}
	m.attachPaper(name, acc)
	return acc, nil
}

// Paper returns the named paper account, creating the default one on first use.
func (m *AccountManager) Paper(name string) (*PaperAccount, error) {
	if name == "" {
		name = DefaultPaperName
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if acc, ok := m.paper[name]; ok {
		return acc, nil
	}
	if name != DefaultPaperName {
		return nil, fmt.Errorf("paper account %q not found", name)
	}
	acc := NewPaperAccount("10000", "0.001")
	if m.db != nil {
		m.db.Create(&models.Account{Name: name, Type: AccountTypePaper, Balance: "10000", Fee: "0.001"})
	}
	m.attachPaper(name, acc)
	return acc, nil
}

// ResetPaper clears positions and restores balance/fee.
func (m *AccountManager) ResetPaper(name, balance, fee string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.paper[name]
	if !ok {
		return fmt.Errorf("paper account %q not found", name)
	}
	b, err := decimal.NewFromString(balance)
	if err != nil {
		return fmt.Errorf("invalid balance: %w", err)
	}
	f, err := decimal.NewFromString(fee)
	if err != nil {
		return fmt.Errorf("invalid fee: %w", err)
	}
	acc.Reset(b, f)
	return nil
}

// TopUp adds amount to the paper account's balance.
func (m *AccountManager) TopUp(name string, amount decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	acc, ok := m.paper[name]
	if !ok {
		return fmt.Errorf("paper account %q not found", name)
	}
	acc.AddBalance(amount)
	return nil
}

// Live returns the live ledger for an exchange ("" = default binance),
// creating an empty one on first use.
func (m *AccountManager) Live(exchangeName string) *LiveAccount {
	if exchangeName == "" {
		exchangeName = DefaultLiveExchange
	}
	exchangeName = strings.ToLower(exchangeName)
	m.mu.Lock()
	defer m.mu.Unlock()
	if acc, ok := m.live[exchangeName]; ok {
		return acc
	}
	acc := NewLiveAccount("0", "0.001")
	m.live[exchangeName] = acc
	return acc
}

// LiveExchanges returns the exchange names that have a live ledger.
func (m *AccountManager) LiveExchanges() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.live))
	for name := range m.live {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// SyncLive refreshes the live ledger for one exchange from real venue balances:
// quote asset becomes the cash balance, every other non-zero asset becomes a position.
func (m *AccountManager) SyncLive(exchangeName string, balances map[string]decimal.Decimal, quote string) {
	acc := m.Live(exchangeName)
	acc.mu.Lock()
	defer acc.mu.Unlock()
	if q, ok := balances[quote]; ok {
		acc.balance = q
	}
	acc.Positions = make(map[string]decimal.Decimal)
	acc.costBasis = make(map[string]decimal.Decimal)
	for asset, amt := range balances {
		if asset != quote && amt.GreaterThan(decimal.Zero) {
			acc.Positions[asset] = amt
		}
	}
}

// List returns all accounts (paper + one live entry per exchange), sorted by name.
func (m *AccountManager) List() []AccountInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]AccountInfo, 0, len(m.paper)+len(m.live))
	for name, acc := range m.paper {
		out = append(out, paperInfo(name, acc))
	}
	for ex, acc := range m.live {
		snap := acc.Snapshot()
		out = append(out, AccountInfo{
			Name:      "live:" + ex,
			Type:      AccountTypeLive,
			Balance:   snap.Balance.String(),
			Fee:       acc.Fee.String(),
			Positions: positionsToString(snap.Positions),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns one account by name: "live" (back-compat = live:binance),
// "live:<exchange>", or a paper account name.
func (m *AccountManager) Get(name string) (AccountInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if name == "live" || name == AccountTypeLive {
		name = "live:" + DefaultLiveExchange
	}
	if strings.HasPrefix(name, "live:") {
		ex := strings.ToLower(name[len("live:"):])
		acc, ok := m.live[ex]
		if !ok {
			return AccountInfo{}, fmt.Errorf("no live account for exchange %q", ex)
		}
		snap := acc.Snapshot()
		return AccountInfo{
			Name: name, Type: AccountTypeLive,
			Balance: snap.Balance.String(), Fee: acc.Fee.String(),
			Positions: positionsToString(snap.Positions),
		}, nil
	}
	if acc, ok := m.paper[name]; ok {
		return paperInfo(name, acc), nil
	}
	return AccountInfo{}, fmt.Errorf("account %q not found", name)
}

func paperInfo(name string, acc *PaperAccount) AccountInfo {
	snap := acc.Snapshot()
	return AccountInfo{
		Name: name, Type: AccountTypePaper,
		Balance: snap.Balance.String(), Fee: acc.Fee.String(),
		Positions: positionsToString(snap.Positions),
	}
}

func positionsToString(positions map[string]decimal.Decimal) map[string]string {
	out := make(map[string]string, len(positions))
	for sym, qty := range positions {
		out[sym] = qty.String()
	}
	return out
}

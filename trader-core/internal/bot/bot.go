package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BotStatus string

const (
	BotCreated  BotStatus = "created"
	BotAttached BotStatus = "attached"
	BotRunning  BotStatus = "running"
)

type BotConfig struct {
	ID       string          `json:"id"`
	Symbol   string          `json:"symbol"`
	Interval engine.Interval `json:"interval"`
	Lookback time.Duration   `json:"lookback"`
	Quantity decimal.Decimal `json:"quantity"`
}

type Bot struct {
	ID       string                    `json:"id"`
	Interval engine.Interval           `json:"interval"`
	Symbol   string                    `json:"symbol"`
	Status   BotStatus                 `json:"status"`
	Started  time.Time                 `json:"started"`
	Strategy strategies.SimpleStrategy `json:"strategy"`
	Quantity decimal.Decimal           `json:"quantity"`

	Engine     engine.ExecutionEngine
	Lookback   time.Duration
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Status != BotAttached {
		return fmt.Errorf("cannot start bot from %s", b.Status)
	}

	if b.Lookback <= 0 {
		return errors.New("lookback must be > 0")
	}

	intervalDur := b.Interval.Duration()
	b.MaxCandles = max(int(b.Lookback/intervalDur), 1)

	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.Started = time.Now()
	b.Status = BotRunning

	go RunBotStrategy(b.ctx, b)
	return nil
}

func (b *Bot) Stop() {
	if b.Status != BotRunning {
		return
	}

	if b.cancel != nil {
		b.cancel()
	}

	b.cancel = nil
	b.ctx = nil
	b.Status = BotAttached
}

func (b *Bot) SetCancel(c context.CancelFunc) {
	b.cancel = c
}

type BotFactory struct {
	Account engine.Account
	Engine  func() engine.ExecutionEngine
}

func (f *BotFactory) NewPaperBot(cfg BotConfig) (*Bot, error) {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}

	b := &Bot{
		ID:     cfg.ID,
		Status: BotCreated,

		Symbol:   cfg.Symbol,
		Quantity: cfg.Quantity,

		Interval: cfg.Interval,
		Lookback: cfg.Lookback,

		Strategy: strategies.SimpleStrategy{},
		Engine:   f.Engine(),
		CandleCh: make(chan models.Candle, 100),
	}

	return b, nil
}

type Runtime struct {
	Account       engine.Account
	BotFactory    *BotFactory
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
	Messenger     *monitoring.Messenger
	Errors        chan string
}

func (rt *Runtime) AttachBot(b *Bot) error {
	if b.Status != BotCreated {
		return fmt.Errorf("cannot attach bot from %s", b.Status)
	}

	// check MarketManager
	if rt.MarketManager == nil || !rt.MarketManager.IsRunning() {
		return fmt.Errorf("cannot attach bot: MarketManager not running")
	}

	// optional: check Binance WS alive
	if !rt.MarketManager.client.IsAlive() {
		return fmt.Errorf("cannot attach bot: Binance client not connected")
	}

	rt.Dispatcher.Subscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Subscribe(b.Symbol, b.Interval)

	b.Status = BotAttached
	return nil
}

func (rt *Runtime) DetachBot(b *Bot) error {
	if b.Status != BotAttached {
		return fmt.Errorf("cannot detach bot from %s", b.Status)
	}

	// check MarketManager
	if rt.MarketManager == nil || !rt.MarketManager.IsRunning() {
		return fmt.Errorf("cannot detach bot: MarketManager not running")
	}

	rt.Dispatcher.Unsubscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Unsubscribe(b.Symbol, b.Interval)

	b.Status = BotCreated
	return nil
}

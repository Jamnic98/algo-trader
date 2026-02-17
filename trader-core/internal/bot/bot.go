package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
)

type BotStatus string

const (
	BotCreated  BotStatus = "created"
	BotAttached BotStatus = "attached"
	BotRunning  BotStatus = "running"
	BotStopped  BotStatus = "stopped"
)

type BotConfig struct {
	ID       string          `json:"id"`
	Symbol   string          `json:"symbol"`
	Interval engine.Interval `json:"interval"`
	Lookback time.Duration   `json:"lookback"`
}

type Bot struct {
	ID         string                    `json:"id"`
	Interval   engine.Interval           `json:"interval"`
	Symbol     string                    `json:"symbol"`
	Status     BotStatus                 `json:"status"`
	Started    time.Time                 `json:"started"`
	Strategy   strategies.SimpleStrategy `json:"strategy"`
	Engine     engine.ExecutionEngine
	Lookback   time.Duration
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Status != BotAttached && b.Status != BotStopped {
		return fmt.Errorf("cannot start bot from %s", b.Status)
	}

	if b.Lookback <= 0 {
		return errors.New("lookback must be >= 0")
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
	if b.cancel != nil {
		b.cancel()
		b.Status = BotStopped
	}
}

func (b *Bot) SetCancel(c context.CancelFunc) {
	b.cancel = c
}

type BotFactory struct {
	PaperAccount *engine.PaperAccount
}

func (f *BotFactory) NewPaperBot(cfg BotConfig) (*Bot, error) {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}

	b := &Bot{
		ID:       cfg.ID,
		Symbol:   cfg.Symbol,
		Interval: cfg.Interval,
		Lookback: cfg.Lookback,
		Strategy: strategies.SimpleStrategy{},
		Engine:   engine.NewPaperExecution(f.PaperAccount),
		CandleCh: make(chan models.Candle, 10),
		Status:   BotCreated,
	}

	return b, nil
}

type Runtime struct {
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
}

func (rt *Runtime) AttachBot(b *Bot) error {
	if b.Status != BotCreated {
		return fmt.Errorf("cannot attach bot from %s", b.Status)
	}

	rt.Dispatcher.Subscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Subscribe(b.Symbol, b.Interval)

	b.Status = BotAttached
	return nil
}

func (rt *Runtime) DetachBot(b *Bot) error {
	if b.Status == BotRunning {
		return fmt.Errorf("cannot detach running bot")
	}

	if b.Status == BotCreated {
		return nil
	}

	rt.Dispatcher.Unsubscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Unsubscribe(b.Symbol, b.Interval)

	b.Status = BotCreated
	return nil
}

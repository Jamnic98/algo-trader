package bot

import (
	"context"
	"errors"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
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
	Status     string                    `json:"status"`
	Started    time.Time                 `json:"started"`
	Strategy   strategies.SimpleStrategy `json:"strategy"`
	Engine     engine.ExecutionEngine
	Lookback   time.Duration
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	Ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Ctx != nil {
		return errors.New("bot already running")
	}

	if b.Lookback <= 0 {
		return errors.New("lookback must be >= 0")
	}

	intervalDur := b.Interval.Duration()
	b.MaxCandles = max(int(b.Lookback/intervalDur), 1)

	b.Ctx, b.cancel = context.WithCancel(context.Background())
	b.Status = "running"
	b.Started = time.Now()

	go RunBotStrategy(b.Ctx, b)
	return nil
}

func (b *Bot) Stop() {
	if b.cancel != nil {
		b.cancel()
		b.Status = "stopped"
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
		Status:   "created",
	}

	return b, nil
}

type Runtime struct {
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
}

func (rt *Runtime) AttachBot(b *Bot) error {
	rt.Dispatcher.Subscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Subscribe(b.Symbol, b.Interval)

	if err := b.Start(); err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) DetatchBot(b *Bot) {
	rt.Dispatcher.Unsubscribe(b.Symbol, b.Interval, b)
	rt.MarketManager.Unsubscribe(b.Symbol, b.Interval)

	b.Stop()
}

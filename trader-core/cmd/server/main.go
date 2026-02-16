package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"
	"trader-core/setup"

	"github.com/google/uuid"
)

const BINANCE_WS_URL = "wss://stream.binance.com:443/ws"

func main() {
	// logger setup
	f, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Tell the logger to write to the file
	log.SetOutput(f)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := setup.GetConfig()
	setup.InitDatabase(cfg)

	server := setup.InitServer()
	wg.Go(func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	})

	binanceClient := binance.NewClient(ctx, BINANCE_WS_URL)
	wg.Go(func() {
		if err := binanceClient.Run(); err != nil {
			log.Fatal(err)
		}
	})

	dispatcher := binance.NewDispatcher()
	marketManager := binance.NewMarketDataManager(binanceClient, dispatcher)
	wg.Go(func() {
		marketManager.Run(ctx)
	})

	account := engine.NewPaperAccount(10000, 0.001)

	// create a bot and start it
	b1 := bot.Bot{
		ID:       uuid.New().String(),
		Symbol:   "BTCUSDT",
		Interval: engine.Interval1m,
		Status:   "running",
		Started:  time.Now(),
		Engine:   engine.NewPaperExecution(account),
		Strategy: strategies.SimpleStrategy{},
		Lookback: 24 * time.Hour,
		CandleCh: make(chan models.Candle, 10),
	}

	// Start bot
	wg.Go(func() {
		if err := b1.Start(); err != nil {
			log.Println(err)
		}
	})

	// wire bot to system
	dispatcher.Subscribe(b1.Symbol, b1.Interval, &b1)
	marketManager.Subscribe(b1.Symbol, b1.Interval)

	// create a bot and start it
	b2 := bot.Bot{
		ID:       uuid.New().String(),
		Symbol:   "SOLUSDT",
		Interval: engine.Interval5m,
		Status:   "running",
		Started:  time.Now(),
		Engine:   engine.NewPaperExecution(account),
		Strategy: strategies.SimpleStrategy{},
		Lookback: 24 * time.Hour,
		CandleCh: make(chan models.Candle, 10),
	}

	// Start bot
	wg.Go(func() {
		if err := b2.Start(); err != nil {
			log.Println(err)
		}
	})

	// wire bot to system
	dispatcher.Subscribe(b2.Symbol, b2.Interval, &b2)
	marketManager.Subscribe(b2.Symbol, b2.Interval)

	wg.Wait()
	<-ctx.Done()
}

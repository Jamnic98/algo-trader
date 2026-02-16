package main

import (
	"context"
	"log"

	"trader-core/internal/api"
	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/setup"
)

const BINANCE_WS_URL = "wss://stream.binance.com:443/ws"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := setup.GetConfig()
	setup.InitDatabase(cfg)

	server := setup.InitServer()
	go func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Binance WS client
	binanceClient := binance.NewClient(ctx, BINANCE_WS_URL)
	if err := binanceClient.Run(); err != nil {
		log.Fatal(err)
	}

	// Dispatcher + MarketManager runtime
	dispatcher := bot.NewDispatcher()
	marketManager := bot.NewMarketDataManager(binanceClient, dispatcher)
	go marketManager.Run(ctx)

	// Paper trading account
	account := engine.NewPaperAccount(10000, 0.001)

	// Shared runtime for bots
	runtime := &bot.Runtime{
		Dispatcher:    dispatcher,
		MarketManager: marketManager,
	}

	// Inject runtime & account into API handlers
	api.InitBotAPI(runtime, account)

	<-ctx.Done()
}

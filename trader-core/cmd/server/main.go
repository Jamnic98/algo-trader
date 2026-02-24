package main

import (
	"context"
	"log"

	"trader-core/internal/api"
	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"
	"trader-core/setup"
)

const binanceWSURL = "wss://stream.binance.com:443/ws"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := setup.GetConfig()
	setup.InitDatabase(cfg)
	server := setup.InitServer(cfg)

	// Run API server
	go func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Binance WS client
	binanceClient := binance.NewClient(ctx, binanceWSURL)
	if err := binanceClient.Run(); err != nil {
		log.Fatal(err)
	}

	// Dispatcher + MarketManager runtime
	dispatcher := bot.NewDispatcher()
	marketManager := bot.NewMarketDataManager(binanceClient, dispatcher)
	go marketManager.Run(ctx)

	// Paper trading account
	account := engine.NewPaperAccount("10000", "0.001")
	botFactory := bot.BotFactory{Account: account, Engine: func() engine.ExecutionEngine {
		return engine.NewPaperExecution(account)
	}}

	messenger := &monitoring.Messenger{
		BotToken: cfg.TelegramKey,
		ChatID:   cfg.TelegramChatID,
		Messages: make(chan string, 10),
		Quit:     make(chan struct{}),
	}

	go messenger.Run()
	defer close(messenger.Quit)

	// Shared runtime for bots
	runtime := &bot.Runtime{
		Account:       account,
		BotFactory:    &botFactory,
		Dispatcher:    dispatcher,
		MarketManager: marketManager,
		Messenger:     messenger,
	}

	// Inject runtime & account into API handlers
	api.InitAccountAPI(runtime)
	api.InitBotAPI(runtime)

	<-ctx.Done()
}

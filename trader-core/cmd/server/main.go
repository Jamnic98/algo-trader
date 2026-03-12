package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trader-core/internal/api"
	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"
	"trader-core/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown on OS signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// config
	cfg := setup.GetConfig()

	// logging
	logPath := ""
	if cfg.Env == "prod" {
		logPath = "bot.log"
	}
	if err := monitoring.SetupLogger(logPath); err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	defer monitoring.ShutdownLogger()

	// database
	db := setup.InitDatabase(cfg)

	// system monitor
	sysmon := monitoring.NewSysMonitor(2 * time.Second)
	go sysmon.Run(ctx)

	// telegram messenger
	messenger := &monitoring.Messenger{
		BotToken: cfg.TelegramKey,
		ChatID:   cfg.TelegramChatID,
		Messages: make(chan string, 10),
		Quit:     make(chan struct{}),
	}
	go messenger.Run()
	defer close(messenger.Quit)

	// market data
	dispatcher := bot.NewDispatcher()
	binanceClient := binance.NewClient(ctx)
	marketManager := bot.NewMarketDataManager(binanceClient, dispatcher)
	go marketManager.Run(ctx)

	// paper trading account and bot factory
	account := engine.NewPaperAccount("10000", "0.001")
	botFactory := bot.BotFactory{
		Account: account,
		Engine: func() engine.ExecutionEngine {
			return engine.NewPaperExecution(account)
		},
	}

	// shared bot runtime
	runtime := &bot.Runtime{
		Account:       account,
		BotFactory:    &botFactory,
		DB:            db,
		Dispatcher:    dispatcher,
		MarketManager: marketManager,
		Messenger:     messenger,
		Errors:        make(chan string, 10),
	}

	// inject dependencies into API handlers
	api.InitAccountAPI(runtime)
	api.InitBotAPI(runtime)
	api.InitDiagnosticsAPI(sysmon)

	// start API server
	server := setup.InitServer(cfg, sysmon)
	go func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("failed to start API server:", err)
		}
	}()

	// connect to binance websocket
	if err := binanceClient.StartWithReconnect(func(msg string) {
		runtime.Errors <- msg
		cancel()
	}); err != nil {
		runtime.Errors <- fmt.Sprintf("failed to connect to Binance: %v", err)
		return
	}
	log.Println("Binance websocket connected")

	// forward errors to log and Telegram
	go func() {
		for {
			select {
			case err := <-runtime.Errors:
				log.Println(err)
				messenger.Notify(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("shutdown complete")
}

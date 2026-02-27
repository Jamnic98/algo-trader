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

const binanceWSURL = "wss://stream.binance.com:443/ws"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Catch OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	cfg := setup.GetConfig()

	// Set up logging
	logPath := ""
	if cfg.Env == "prod" {
		logPath = "bot.log"
	}
	if err := monitoring.SetupLogger(logPath); err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	defer monitoring.ShutdownLogger()

	// Initialize backend
	setup.InitDatabase(cfg)
	server := setup.InitServer(cfg)

	// Run API server
	go func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start API server:", err)
		}
	}()

	// Centralized errors channel
	errors := make(chan string, 20)

	// Set up Messenger for Telegram notifications
	messenger := &monitoring.Messenger{
		BotToken: cfg.TelegramKey,
		ChatID:   cfg.TelegramChatID,
		Messages: make(chan string, 10),
		Quit:     make(chan struct{}),
	}
	go messenger.Run()
	defer close(messenger.Quit)

	// Listener: send all errors to log + Telegram
	go func() {
		for errMsg := range errors {
			log.Println(errMsg)
			messenger.Notify(errMsg)
		}
	}()

	// Dispatcher + MarketManager
	dispatcher := bot.NewDispatcher()
	binanceClient := binance.NewClient(ctx, binanceWSURL)
	marketManager := bot.NewMarketDataManager(binanceClient, dispatcher)
	go marketManager.Run(ctx)

	// Paper trading account
	account := engine.NewPaperAccount("10000", "0.001")

	// Bot factory
	botFactory := bot.BotFactory{
		Account: account,
		Engine: func() engine.ExecutionEngine {
			return engine.NewPaperExecution(account)
		},
	}

	// Shared runtime for bots
	runtime := &bot.Runtime{
		Account:       account,
		BotFactory:    &botFactory,
		Dispatcher:    dispatcher,
		MarketManager: marketManager,
		Messenger:     messenger,
		Errors:        errors,
	}

	// Connect to binance websocket
	go func() {
		log.Println("Connecting to Binance websocket...")
		if err := binanceClient.Run(); err != nil {
			runtime.Errors <- fmt.Sprintf("Failed to connect to Binance: %v", err)
			return
		}
		log.Println("Binance websocket connection established")
	}()

	// Run reconnect routine for binance websocket
	go func() {
		const maxReconnects = 3
		const reconnectDelay = 2 * time.Second
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		reconnectAttempts := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !binanceClient.Started() {
					continue // do nothing until ready
				}

				if binanceClient.IsAlive() {
					reconnectAttempts = 0
					continue
				}

				reconnectAttempts++
				if reconnectAttempts > maxReconnects {
					runtime.Errors <- "CRITICAL: Binance WS could not reconnect after max attempts"
					cancel()
					return
				}

				log.Printf("WARNING: Binance WS disconnected, attempt %d/%d to reconnect", reconnectAttempts, maxReconnects)
				if err := binanceClient.Run(); err != nil {
					runtime.Errors <- fmt.Sprintf("WARNING: reconnect failed: %v", err)
				} else {
					log.Println("Binance websocket connection established")
					reconnectAttempts = 0
				}

				time.Sleep(reconnectDelay)
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-runtime.Errors:
				messenger.Notify(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Inject runtime into API handlers
	api.InitAccountAPI(runtime)
	api.InitBotAPI(runtime)

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Main context canceled, shutting down")
}

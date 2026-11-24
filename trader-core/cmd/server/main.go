package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "trader-core/internal/alpaca"
	"trader-core/internal/api"
	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"
	"trader-core/internal/monitoring"
	"trader-core/setup"

	"github.com/shopspring/decimal"
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
		logPath = "logs/bot.log"
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

	// market data: one shared Binance WS client + a broker per registered venue
	// (binance, alpaca, ...) for historical klines, streaming, balances and orders
	dispatcher := bot.NewDispatcher()
	binanceClient := binance.NewClient(ctx)
	brokers := map[string]exchange.Broker{}
	for _, name := range exchange.Names() {
		br, err := exchange.New(name, exchange.ConfigFromEnv(name))
		if err != nil {
			log.Printf("exchange %q unavailable: %v", name, err)
			continue
		}
		brokers[name] = br
		log.Printf("exchange %q ready (paper=%v)", name, exchange.ConfigFromEnv(name).Paper)
	}
	marketManager := bot.NewMarketDataManager(binanceClient, dispatcher, brokers)
	go marketManager.Run(ctx)

	// default order placer (binance); live bots pick their own venue's broker
	broker := brokers["binance"]
	var placer engine.OrderPlacer = binance.SpotPlacer{}
	if broker != nil {
		placer = broker // a Broker satisfies engine.OrderPlacer
	}

	// account manager: default paper account + live ledger synced from the broker
	liveAccount := engine.NewLiveAccount("0", "0.001")
	var balances map[string]decimal.Decimal
	if broker != nil {
		if b, err := broker.Balances(); err != nil {
			log.Println("could not fetch exchange balances (live ledger starts at 0):", err)
		} else {
			balances = b
			if usdt, ok := b["USDT"]; ok {
				log.Println("live USDT balance:", usdt.String())
			}
		}
	}
	accountManager := engine.NewAccountManager(db, liveAccount)
	if balances != nil {
		accountManager.SyncLive(balances, "USDT")
	}
	account, err := accountManager.Paper(engine.DefaultPaperName) // "default": 10000 USDT paper
	if err != nil {
		log.Fatal(err)
	}
	botFactory := bot.BotFactory{
		Account: account,
		Engine: func() engine.ExecutionEngine {
			return engine.NewPaperExecutor(account)
		},
		LiveAccount: liveAccount,
		Placer:      placer,
	}

	// shared bot runtime
	runtime := &bot.Runtime{
		Account:       account,
		Accounts:      accountManager,
		BotFactory:    &botFactory,
		DB:            db,
		Dispatcher:    dispatcher,
		MarketManager: marketManager,
		Messenger:     messenger,
		Errors:        make(chan string, 10),
		Brokers:       brokers,
	}

	// inject dependencies into API handlers
	api.InitAccountAPI(runtime)
	api.InitBotAPI(runtime)
	api.InitDiagnosticsAPI(sysmon)

	// start API server
	server := setup.InitServer(cfg)
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

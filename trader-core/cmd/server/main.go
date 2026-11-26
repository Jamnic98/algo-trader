package main

import (
	"context"
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

	// account manager: default paper account + one live ledger per exchange,
	// each synced from that venue's real balances
	liveAccount := engine.NewLiveAccount("0", "0.001")
	accountManager := engine.NewAccountManager(db, liveAccount)
	for name, br := range brokers {
		bals, err := br.Balances()
		if err != nil {
			log.Printf("could not fetch %s balances (live:%s starts at 0): %v", name, name, err)
			continue
		}
		accountManager.SyncLive(name, bals, exchange.QuoteAssetFor(name))
		log.Printf("live:%s synced (%s balance: %s)", name, exchange.QuoteAssetFor(name), bals[exchange.QuoteAssetFor(name)].String())
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

	// connect to binance websocket; a dial failure must NOT take the API down
	// (alpaca bots and REST keep working) — retry in the background instead
	go binanceClient.StartWithRetry(ctx, 5*time.Second, func(msg string) {
		runtime.Errors <- msg
	})
	log.Println("Binance websocket connecting (retries in background on failure)")

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

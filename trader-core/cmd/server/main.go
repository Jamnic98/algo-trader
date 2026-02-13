package main

import (
	"context"
	"log"

	// "trader-core/internal/redis"
	"trader-core/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := setup.GetConfig()
	setup.InitDatabase(cfg)

	// rdb := redis.NewClient()
	server := setup.InitServer()

	go func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	/* 	// Binance
	   	client := binance.NewClient(ctx, "wss://stream.binance.com:443/ws")
	   	go func() {
	   		if err := client.Run(); err != nil {
	   			log.Fatal(err)
	   		}
	   	}()

	   	client.Subscribe("btcusdt@aggTrade", "btcusdt@depth") */

	// wait for shutdown signal
	<-ctx.Done()
}

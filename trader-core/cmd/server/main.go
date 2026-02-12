package main

import (
	"context"
	"log"

	"trader-core/internal/binance"
	"trader-core/internal/redis"
	"trader-core/setup"
)


func main() {
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  cfg := setup.GetConfig()
  setup.InitDatabase(cfg)
  rdb := redis.NewClient()
  server := setup.InitServer()

  go func() {
      if err := server.Run(":" + cfg.Port); err != nil {
          log.Fatal("Failed to start server:", err)
      }
  }()

  client := binance.NewClient(ctx, "wss://stream.binance.com:443/ws")
  go func() {
    if err := client.Run(); err != nil {
      log.Fatal(err)
    }
  }()

  client.Subscribe("btcusdt@aggTrade", "btcusdt@depth")

  for msg := range client.Messages() {
    // Process and optionally push to Redis
    // log.Println(string(msg))
    err := rdb.Set(ctx, "last_message", msg, 0).Err()
    if err != nil {
        log.Println("Redis set failed:", err)
    }
  }

  <-ctx.Done() // wait for shutdown signal
}



/* // Connect to Binance WebSocket
binanceConnection := binance.ConnectToBinanceWs()

// Subscribe
subMsg := map[string]any{
  "method": "SUBSCRIBE",
  "params": []string{"btcusdt@aggTrade", "btcusdt@depth"},
  "id":     1,
}
if err := binanceConnection.WriteJSON(subMsg); err != nil {
  log.Fatal("WebSocket subscribe failed:", err)
}

// Start reading messages in a goroutine
wg.Go(func() {
  defer binanceConnection.Close()
  for {
    _, msg, err := binanceConnection.ReadMessage()
    if err != nil {
      log.Println("WebSocket read error:", err)
      return
    }
    log.Println("Received:", string(msg))
  }
}) */

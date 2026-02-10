package main

import (
	"log"
	"sync"

	"api/setup"
)

func main() {
  // Initialisation
  var wg = sync.WaitGroup{}
  
  cfg := setup.GetConfig()
  setup.InitDatabase(cfg)
  server := setup.InitServer()

  // Start server
  wg.Add(1)
  go func() {
    defer wg.Done()
    if err := server.Run(":" + cfg.Port); err != nil {
      log.Fatal("Failed to start server:", err)
    }
  }()
  log.Printf("Server running on port %s", cfg.Port)

  /*
  // Connect to Binance WebSocket
  url := "wss://stream.binance.com:443/ws"
  conn, _, err := websocket.DefaultDialer.Dial(url, nil)
  if err != nil {
    log.Fatal("Error connecting to WebSocket:", err)
  }
  log.Printf("Connected to Binance WebSocket!")

  // Subscribe
  subMsg := map[string]interface{}{
    "method": "SUBSCRIBE",
    "params": []string{"btcusdt@aggTrade", "btcusdt@depth"},
    "id":     1,
  }
  if err := conn.WriteJSON(subMsg); err != nil {
    log.Fatal("WebSocket subscribe failed:", err)
  }

  // Start reading messages in a goroutine
  wg.Add(1)
  go func() {
    defer wg.Done()
    defer conn.Close()
    for {
      _, msg, err := conn.ReadMessage()
      if err != nil {
        log.Println("WebSocket read error:", err)
        return
      }
      log.Println("Received:", string(msg))
    }
  }()
 */

  wg.Wait()
}

# Trader Core

A modular, event-driven crypto trading bot engine written in Go.

This project is designed to run multiple algorithmic trading bots against live market data (Binance WebSocket),
with clean separation between market data ingestion, candle dispatching, strategy execution, and API control.

---

## High-level architecture

The system is split into a few core components:

### Binance Client
- Maintains a WebSocket connection to Binance
- Receives raw kline (candle) events
- Pushes messages into an internal channel

### MarketDataManager
- Consumes raw WebSocket messages
- Parses kline events
- Filters **only closed candles**
- Fans out candles to subscribed bots via the Dispatcher

### Dispatcher
- Maintains subscriptions per symbol + interval
- Sends candles to each bot's CandleCh
- Non-blocking fan-out (drops if a bot is slow)

### Bot
- Owns strategy execution
- Maintains internal candle buffer (lookback-based)
- Executes strategy logic on each closed candle
- Can be started, stopped, attached, or detached at runtime

### Runtime
- Glue layer between Dispatcher, MarketDataManager, and Bots
- Responsible for attaching and detaching bots from live feeds

### API (Gin)
- Create, start, stop, delete bots via HTTP
- Returns safe DTOs (no channels or internal state leaks)

### Frontend (Vite + Tailwind)
- Web-based control panel for managing bots
- Communicates with the Go API over HTTP
- Displays bot state, configuration, and lifecycle
- Designed to be fast, minimal, and extendable

---

## Bot lifecycle and states

Bots intentionally separate **market attachment** from **strategy execution**.

### States

- `created`  
  Bot exists but is not connected to market data.

- `attached`  
  Bot is subscribed to market data and receiving candles, but strategy is not running.

- `running`  
  Bot is attached **and** actively executing its strategy.

- `stopped`  
  Strategy loop is stopped, but the bot may still be attached.

### Key ideas

- **Attach** means: subscribed to feeds, consuming resources.
- **Run** means: executing strategy and placing trades.
- **Stop** means: strategy paused, market feed may still be attached.
- **Detach** means: fully unsubscribed and idle.

This separation allows bots to be paused and resumed without resubscribing to Binance.

---

## Concurrency model

- WebSocket reader runs in its own goroutine
- MarketDataManager runs in a dedicated goroutine
- Each bot strategy runs in its own goroutine
- Context cancellation is used for clean shutdowns
- Channels are non-blocking where fan-out occurs

---

## Frontend setup (Vite + Tailwind)

The frontend lives in its own directory and acts purely as a UI layer.

### Responsibilities

- Create bots
- Start and stop bots
- View bot status and configuration
- Inspect running state in real time

## Design goals

- Deterministic behaviour
- Clear lifecycle boundaries
- Safe concurrency
- Easy extensibility for new strategies
- API-driven control plane

---

## Notes

This project intentionally avoids:
- Global mutable state where possible
- Blocking fan-out
- Implicit goroutine lifetimes

Everything should start, stop, and clean up explicitly.

___


### Running the app with docker compose


``` bash
docker compose -f docker-compose.prod.yml up --build
```
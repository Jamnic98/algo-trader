# algo-trader

Monorepo: Go trading core + React dashboard + Python backtester.

- `trader-core/` — Go API (Gin) and live/paper bot engine. Serves `/api/*` on `:8080`.
- `dashboard/` — React + Vite frontend on `:5173` (dev).
- `research/` — Python (FastAPI) backtester service on `:8000`. Reads saved strategies
  from Postgres and runs them against historical Binance klines.
- `shared/strategies.json` — **single source of truth for strategy definitions**
  (templates, param schemas/defaults, behavior test vectors). `research` (Python)
  loads it at runtime; `trader-core` (Go) embeds it via the generated
  `trader-core/internal/strategies/gen_schema.go`. After editing the JSON, run
  `python3 scripts/generate_strategies.py` (CI runs `--check` and fails if stale).
  `/api/strategies/schemas` is served from it, and both test suites execute its
  test vectors, so the Go and Python engines can't silently drift apart.

To add a strategy on an existing template: add an entry to `shared/strategies.json`
and re-run the generator — no Go/Python code changes needed.

### Combining strategies (composites)

Any saved strategy — primitive **or composite** — can be used as a sub-strategy node
in another composite (`{"threshold": 0.6, "nodes": [{"strategy_slug": ..., "weight": ...,
"params": {...}}]}`), nested arbitrarily deep. Both engines resolve recursively.
Create them from the dashboard wizard (`/strategies/create`), which lists all saved
strategies as node options.

### Custom / ML strategies (the `http` template)

For logic that isn't a built-in template (e.g. a machine-learning model), add a
primitive with `"template": "http"` and a `url` param pointing at your inference
service. On every candle, both engines POST `{"closes": [...window...], "params": {...}}`
to that URL and expect `{"signal": "BUY"|"SELL"|"HOLD", "quantity": "..." (optional)}`.
Unreachable/invalid responses safely return HOLD. See the seeded `ml_example`
strategy and `scripts/ml_example_service.py` (toy model server on :8001) for a
working example:

```bash
python3 scripts/ml_example_service.py   # model endpoint on 127.0.0.1:8001
# then backtest "ml_example" from the dashboard, or combine it into a composite
```

## Strategy → backtest flow

1. Create a strategy in the dashboard (`/strategies/create`) → `POST /api/strategies`
   stores it in the `strategies` table (slug, display name, composite config).
2. On a strategy's page, run a backtest → `POST /api/backtests` (trader-core) validates
   the strategy and proxies to the backtester (`BACKTESTER_URL`, e.g.
   `http://research:8000`).
3. The backtester loads the strategy from the DB, fetches klines from Binance REST
   (`GET /api/v3/klines`, paginated), replays them through a paper account and returns
   stats, equity curve and trades.

## Risk management & panic

Every bot carries a `risk` JSON object (stored on the bot's `risk` jsonb column;
defaults applied when omitted). Set it in the bot form's **Risk** section
(dashboard) or as `"risk": {...}` in `POST /api/bots`:

```json
{
  "max_position_pct": "25",     // cap a single BUY notional as % of account balance (default "100" = no cap)
  "max_daily_loss_pct": "5",    // kill switch: equity drop vs day-start equity; "0" = off
  "max_drawdown_pct": "10",     // kill switch: equity drawdown from running peak; "0" = off
  "cooldown_losses": 3,         // consecutive losing SELLs before BUYs pause; 0 = off
  "cooldown_minutes": 60        // pause length
}
```

Enforcement happens in the bot runner on every candle: position cap before the
fill, kill-switch checks on equity (a trip stops/detaches the bot and sends a
Telegram alert), cooldown blocks new BUYs, and each fill's realized pnl (tracked
via per-position cost basis) feeds the cooldown. `GET /api/bots` returns each
bot's stored `risk`.

**Panic button / endpoint:** the Bots page has a *Panic — stop all* button
(confirm, then optional flatten). API equivalent: `POST /api/panic` with
`{"flatten": true}` stops all running bots and, with `flatten`, market-sells
every open position at the last candle close (real orders for live bots).
Response: `{stopped, flattened, errors}`.

## Reliability: disconnects, reconnects, restarts

- The Binance WS client **retries forever in the background** (`StartWithRetry`)
  — a failed dial or fatal disconnect no longer kills the process, and the REST
  API / other venues keep working.
- Subscriptions are **replayed on every (re)connect**, and any closed candles
  missed during the gap are **backfilled** (`Broker.Klines` / `FetchKlinesRange`
  from the last seen close; Binance via a client `OnReconnect` hook), so bots
  never trade on a stale window after a reconnect.
- **Paper positions are persisted** (`accounts.positions` jsonb) and restored on
  boot, so an open position survives a restart and can still be sold. Each
  exchange gets its own live ledger (`live:binance`, `live:alpaca`, …) with the
  venue's quote asset (`QuoteAssetFor`: USDT/USD), rehydrated via
  `SyncLive(exchange, balances, quote)` at boot, on live-bot attach, and via
  `POST /api/account/live/sync` (`?exchange=` for one venue, no param for all);
  `GET /api/account/live?exchange=` reads one ledger (`live` alone still means
  Binance for back-compat).
- **Live-order safety:** every order carries a client order id (`botID-nanos`);
  on a transport error or 5xx the broker reconciles by that id
  (Binance `origClientOrderId`, Alpaca `client_order_id`) and returns the fill if
  it actually executed — a dropped response can't double-fill or drift the
  ledger. Venue trading rules (`Broker.Filters`: lot size / tick / min notional)
  are fetched on bot attach and validated client-side before each order.

## Run locally

```bash
docker compose -f docker-compose.local.yml up --build
```

Required env in `.env.local`: `SERVER_API_KEY`, `DB_DSN`, and for live trading
`BINANCE_API_KEY` / `BINANCE_SECRET_KEY` (never commit these).

## Multi-exchange (Alpaca)

Bots carry an `exchange` field (`binance` by default). Venues register themselves
in the `internal/exchange` registry (`internal/binance`, `internal/alpaca`), which
provides klines, websocket streaming, balances and market orders behind one
`exchange.Broker` interface. The backtester accepts `"exchange": "binance" |
"alpaca"` on `POST /api/backtests` / `POST /backtest` (Alpaca bars come from
`data.alpaca.markets`: stocks via `/v2/stocks/{sym}/bars`, crypto via
`/v1beta3/crypto/us/{sym}/bars`).

Alpaca env (paper trading API by default, live when `ALPACA_PAPER` is unset/false):

```
ALPACA_API_KEY=...
ALPACA_SECRET=...
ALPACA_PAPER=true            # use https://paper-api.alpaca.markets for orders
# optional overrides:
ALPACA_BASE_URL=...          # trading REST (default paper/live api.alpaca.markets)
ALPACA_DATA_URL=...          # market data (default https://data.alpaca.markets)
ALPACA_WS_URL=...            # market data websocket
```

The research service reads `ALPACA_DATA_URL` / `ALPACA_API_KEY` / `ALPACA_SECRET`
for backtests; the trader-core reads the same vars (compose passes `.env.prod`
through to both).

## Run in prod

```bash
docker compose -f docker-compose.prod.yml up --build
```

# algo-trader

Monorepo: Go trading core + React dashboard + Python backtester.

- `trader-core/` — Go API (Gin) and live/paper bot engine. Serves `/api/*` on `:8080`.
- `dashboard/` — React + Vite frontend on `:5173` (dev).
- `research/` — Python (FastAPI) backtester service on `:8000`. Reads saved strategies
  from Postgres and runs them against historical Binance klines.
- `shared/strategies.json` — **single source of truth for strategy definitions**
  (templates, param schemas/defaults, behavior test vectors). Both `trader-core` and
  `research` load it at startup (`STRATEGIES_SCHEMA_PATH` to override); the
  `/api/strategies/schemas` endpoint is served from it, and both test suites execute
  its test vectors, so the Go and Python engines can't silently drift apart.

## Strategy → backtest flow

1. Create a strategy in the dashboard (`/strategies/create`) → `POST /api/strategies`
   stores it in the `strategies` table (slug, display name, composite config).
2. On a strategy's page, run a backtest → `POST /api/backtests` (trader-core) validates
   the strategy and proxies to the backtester (`BACKTESTER_URL`, e.g.
   `http://research:8000`).
3. The backtester loads the strategy from the DB, fetches klines from Binance REST
   (`GET /api/v3/klines`, paginated), replays them through a paper account and returns
   stats, equity curve and trades.

## Run locally

```bash
docker compose -f docker-compose.local.yml up --build
```

Required env in `.env.local`: `SERVER_API_KEY`, `DB_DSN`, and for live trading
`BINANCE_API_KEY` / `BINANCE_SECRET_KEY` (never commit these).

## Run in prod

```bash
docker compose -f docker-compose.prod.yml up --build
```

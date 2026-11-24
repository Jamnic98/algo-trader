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

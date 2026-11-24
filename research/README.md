# research — Python backtester service

FastAPI service that runs saved strategies against historical Binance klines.

## Flow

1. The dashboard creates a strategy via `POST /api/strategies` (trader-core), which
   stores it in the Postgres `strategies` table.
2. `POST /api/backtests` on trader-core validates the strategy and proxies the job to
   this service (`BACKTESTER_URL`, default `http://localhost:8000`).
3. This service loads the strategy row from Postgres (`DB_DSN`), resolves it
   (primitives: `simple`, `simple_dca`, `smart_dca`; composites: weighted threshold
   composer — mirroring the Go `internal/strategies` semantics), fetches klines from
   the Binance REST API (`/api/v3/klines`, paginated by start/end time) and replays
   them through a paper account.
4. Returns stats (return, max drawdown, win rate, profit factor, sharpe), the equity
   curve and the trade list.

## API

- `GET /health` → `{"status": "ok"}`
- `POST /backtest`

```json
{
  "strategy_slug": "simple",
  "symbol": "BTCUSDT",
  "interval": "1h",
  "start": "2024-01-01",
  "end": "2024-06-01",
  "initial_balance": "10000",
  "fee_rate": "0.001",
  "quantity": "0.001"
}
```

`start`/`end` accept unix ms, RFC3339 or `YYYY-MM-DD` (UTC). `quantity` is optional —
when omitted, buys are all-in and sells close the full position.

## Run locally

```bash
pip install -r requirements.txt
DB_DSN=postgres://... uvicorn app.main:app --host 0.0.0.0 --port 8000
```

Or via docker compose (see root `docker-compose.local.yml`, service `research`).

## Dev checks

```bash
./scripts/dev-check.sh   # black, ruff, mypy
python -m pytest         # tests (Binance + DB are mocked)
```

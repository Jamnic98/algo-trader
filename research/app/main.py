from decimal import Decimal, InvalidOperation

import httpx
from fastapi import Depends, FastAPI, HTTPException

from . import binance, config, engine, models
from .store import PostgresStrategyStore, StrategyRow, StrategyStore
from .strategies import resolve
from .timeutil import parse_time_ms

app = FastAPI(title="algo-trader backtester")

_store: StrategyStore | None = None


def get_store() -> StrategyStore:
    """Lazy singleton so tests can inject a StaticStrategyStore via set_store()."""
    global _store
    if _store is None:
        if not config.DB_DSN:
            raise HTTPException(status_code=500, detail="DB_DSN not configured")
        _store = PostgresStrategyStore(config.DB_DSN)
    return _store


def set_store(store: StrategyStore) -> None:
    global _store
    _store = store


def get_http_client() -> httpx.AsyncClient:
    return httpx.AsyncClient(base_url=config.BINANCE_REST_BASE, timeout=30.0)


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.post("/backtest", response_model=models.BacktestResult)
async def run_backtest(
    req: models.BacktestRequest,
    store: StrategyStore = Depends(get_store),
    client: httpx.AsyncClient = Depends(get_http_client),
) -> models.BacktestResult:
    row: StrategyRow | None = store.get(req.strategy_slug)
    if row is None:
        raise HTTPException(status_code=404, detail="strategy not found")

    try:
        strategy = resolve(row, None, store)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc

    try:
        start_ms = parse_time_ms(req.start)
        end_ms = parse_time_ms(req.end)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    if end_ms <= start_ms:
        raise HTTPException(status_code=400, detail="end must be after start")

    try:
        initial_balance = Decimal(req.initial_balance)
        fee_rate = Decimal(req.fee_rate)
    except InvalidOperation as exc:
        raise HTTPException(status_code=400, detail=f"invalid decimal: {exc}") from exc

    try:
        candles = await binance.fetch_klines(
            client, req.symbol, req.interval, start_ms, end_ms
        )
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except httpx.HTTPError as exc:
        raise HTTPException(
            status_code=502, detail=f"binance request failed: {exc}"
        ) from exc

    if not candles:
        raise HTTPException(status_code=404, detail="no klines in the requested range")

    result = engine.run_backtest(
        candles,
        strategy,
        req.symbol,
        initial_balance,
        fee_rate,
        req.quantity,
    )

    return models.BacktestResult(
        strategy_slug=req.strategy_slug,
        symbol=req.symbol,
        interval=req.interval,
        start=start_ms,
        end=end_ms,
        stats=result["stats"],
        equity_curve=result["equity_curve"],
        trades=result["trades"],
    )

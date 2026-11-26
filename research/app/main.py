import itertools
from decimal import Decimal, InvalidOperation
from typing import Any

import httpx
from fastapi import Depends, FastAPI, HTTPException

from . import alpaca, binance, config, engine, models
from .binance import Candle
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


def get_alpaca_client() -> httpx.AsyncClient:
    headers = {}
    if config.ALPACA_API_KEY and config.ALPACA_SECRET:
        headers = {
            "APCA-API-KEY-ID": config.ALPACA_API_KEY,
            "APCA-API-SECRET-KEY": config.ALPACA_SECRET,
        }
    return httpx.AsyncClient(
        base_url=config.ALPACA_DATA_URL, timeout=30.0, headers=headers
    )


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


def _load_strategy(store: StrategyStore, slug: str) -> StrategyRow:
    row = store.get(slug)
    if row is None:
        raise HTTPException(status_code=404, detail="strategy not found")
    return row


def _parse_common(
    req: models.BacktestRequest,
) -> tuple[int, int, Decimal, Decimal, Decimal]:
    """Validate times/decimals shared by /backtest and /backtest/sweep."""
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
        slippage_bps = Decimal(req.slippage_bps)
    except InvalidOperation as exc:
        raise HTTPException(status_code=400, detail=f"invalid decimal: {exc}") from exc
    if slippage_bps < 0:
        raise HTTPException(status_code=400, detail="slippage_bps must be >= 0")
    return start_ms, end_ms, initial_balance, fee_rate, slippage_bps


async def _fetch_candles(
    req: models.BacktestRequest,
    client: httpx.AsyncClient,
    alpaca_client: httpx.AsyncClient,
) -> tuple[str, dict[str, list[Candle]]]:
    """Fetch candles for every requested symbol, keyed by normalized symbol."""
    raw_symbols = req.symbols or ([req.symbol] if req.symbol else [])
    if not raw_symbols:
        raise HTTPException(status_code=400, detail="symbol or symbols is required")
    start_ms = parse_time_ms(req.start)  # already validated; cheap to re-parse
    end_ms = parse_time_ms(req.end)

    exchange_name = req.exchange.lower()
    candle_map: dict[str, list[Candle]] = {}
    try:
        for raw in raw_symbols:
            if exchange_name == "binance":
                symbol = binance.normalize_symbol(raw)
                candles = await binance.fetch_klines(
                    client, symbol, req.interval, start_ms, end_ms
                )
            elif exchange_name == "alpaca":
                symbol = alpaca.normalize_symbol(raw)
                candles = await alpaca.fetch_bars(
                    alpaca_client, symbol, req.interval, start_ms, end_ms
                )
            else:
                raise HTTPException(
                    status_code=400,
                    detail=f"unsupported exchange: {req.exchange!r} (supported: binance, alpaca)",
                )
            if candles:
                candle_map[symbol] = candles
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except httpx.HTTPError as exc:
        raise HTTPException(
            status_code=502, detail=f"{exchange_name} request failed: {exc}"
        ) from exc

    if not candle_map:
        raise HTTPException(status_code=404, detail="no klines in the requested range")
    return exchange_name, candle_map


@app.post("/backtest", response_model=models.BacktestResult)
async def run_backtest(
    req: models.BacktestRequest,
    store: StrategyStore = Depends(get_store),
    client: httpx.AsyncClient = Depends(get_http_client),
    alpaca_client: httpx.AsyncClient = Depends(get_alpaca_client),
) -> models.BacktestResult:
    row = _load_strategy(store, req.strategy_slug)
    start_ms, end_ms, initial_balance, fee_rate, slippage_bps = _parse_common(req)
    exchange_name, candle_map = await _fetch_candles(req, client, alpaca_client)

    # one strategy instance per symbol: strategies keep internal state, so
    # symbols must not share an instance in a multi-symbol run
    try:
        strategies = {sym: resolve(row, None, store) for sym in candle_map}
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc

    result = engine.run_portfolio(
        candle_map,
        strategies,
        initial_balance,
        fee_rate,
        req.quantity,
        slippage_bps=slippage_bps,
    )

    return models.BacktestResult(
        strategy_slug=req.strategy_slug,
        exchange=exchange_name,
        symbol=result["symbols"][0],
        symbols=result["symbols"],
        interval=req.interval,
        start=start_ms,
        end=end_ms,
        stats=result["stats"],
        equity_curve=result["equity_curve"],
        trades=result["trades"],
    )


_SORT_KEYS = {
    "total_return_pct",
    "sharpe",
    "profit_factor",
    "max_drawdown_pct",
    "num_trades",
}


def _sort_results(items: list[models.SweepResultItem], sort_by: str) -> None:
    if sort_by not in _SORT_KEYS:
        raise HTTPException(
            status_code=400,
            detail=f"invalid sort_by: {sort_by!r} (allowed: {sorted(_SORT_KEYS)})",
        )
    ascending = sort_by == "max_drawdown_pct"

    def key(item: models.SweepResultItem) -> Decimal:
        raw = getattr(item.stats, sort_by)
        if raw is None:
            return Decimal("-inf") if not ascending else Decimal("inf")
        try:
            return Decimal(str(raw))
        except InvalidOperation:
            return Decimal("inf")  # e.g. profit_factor "inf"

    items.sort(key=key, reverse=not ascending)


@app.post("/backtest/sweep", response_model=models.SweepResponse)
async def run_sweep(
    req: models.SweepRequest,
    store: StrategyStore = Depends(get_store),
    client: httpx.AsyncClient = Depends(get_http_client),
    alpaca_client: httpx.AsyncClient = Depends(get_alpaca_client),
) -> models.SweepResponse:
    """Run one backtest per combination of `param_grid` (cartesian product).

    Market data is fetched once and reused for every combination. Overrides
    apply to the strategy's top-level config (see strategies.resolve).
    """
    row = _load_strategy(store, req.strategy_slug)
    start_ms, end_ms, initial_balance, fee_rate, slippage_bps = _parse_common(req)
    exchange_name, candle_map = await _fetch_candles(req, client, alpaca_client)

    if not req.param_grid:
        raise HTTPException(status_code=400, detail="param_grid must not be empty")
    keys = sorted(req.param_grid)
    for k, values in req.param_grid.items():
        if not values:
            raise HTTPException(
                status_code=400, detail=f"param_grid[{k!r}] has no values"
            )
    combos = list(itertools.product(*(req.param_grid[k] for k in keys)))
    max_combos = req.max_combos if req.max_combos > 0 else 200
    if len(combos) > max_combos:
        raise HTTPException(
            status_code=400,
            detail=f"param grid too large: {len(combos)} combos > max_combos {max_combos}",
        )

    results: list[models.SweepResultItem] = []
    for combo in combos:
        overrides: dict[str, Any] = dict(zip(keys, combo))
        try:
            strategies = {sym: resolve(row, overrides, store) for sym in candle_map}
        except ValueError as exc:
            raise HTTPException(status_code=400, detail=str(exc)) from exc
        outcome = engine.run_portfolio(
            candle_map,
            strategies,
            initial_balance,
            fee_rate,
            req.quantity,
            slippage_bps=slippage_bps,
        )
        results.append(
            models.SweepResultItem(
                params=overrides,
                stats=outcome["stats"],
                equity_curve=outcome["equity_curve"],
                trades=outcome["trades"] if req.include_trades else [],
            )
        )

    _sort_results(results, req.sort_by)
    return models.SweepResponse(
        strategy_slug=req.strategy_slug,
        exchange=exchange_name,
        symbols=sorted(candle_map),
        interval=req.interval,
        start=start_ms,
        end=end_ms,
        count=len(results),
        sort_by=req.sort_by,
        results=results,
    )

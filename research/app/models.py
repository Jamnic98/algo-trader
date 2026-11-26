from typing import Any

from pydantic import BaseModel


class BacktestRequest(BaseModel):
    strategy_slug: str
    exchange: str = "binance"  # venue for market data: "binance" or "alpaca"
    symbol: str | None = None  # single venue-native symbol (BTCUSDT, BTC/USD, AAPL)
    symbols: list[str] | None = None  # multi-symbol portfolio run (overrides symbol)
    interval: str
    start: str | int  # unix ms, RFC3339 or YYYY-MM-DD
    end: str | int
    initial_balance: str = "10000"
    fee_rate: str = "0.001"
    slippage_bps: str = "0"  # adverse slippage on fills, in basis points
    quantity: str | None = None  # fixed qty when the strategy doesn't dictate one


class TradeOut(BaseModel):
    time: int  # candle open time, ms
    side: str  # BUY / SELL
    price: str
    qty: str
    fee: str
    balance: str  # cash balance after the fill
    pnl: str | None = None  # realized pnl on sells
    symbol: str | None = None  # which symbol this fill belongs to


class StatsOut(BaseModel):
    initial_balance: str
    final_balance: str
    total_return_pct: str
    max_drawdown_pct: str
    num_trades: int
    win_rate: str  # % of sells with positive realized pnl
    profit_factor: str | None  # gross profit / gross loss; None when no losses
    sharpe: (
        str | None
    )  # annualized-ish per-candle sharpe (scaled by sqrt(n)); None if undefined


class EquityPoint(BaseModel):
    time: int
    equity: str


class BacktestResult(BaseModel):
    strategy_slug: str
    exchange: str = "binance"
    symbol: str  # first symbol (back-compat)
    symbols: list[str] = []  # all symbols in the run
    interval: str
    start: int
    end: int
    stats: StatsOut
    equity_curve: list[EquityPoint]
    trades: list[TradeOut]


class SweepRequest(BacktestRequest):
    """Run the same backtest over a cartesian product of param overrides.

    Overrides are applied to the strategy's top-level config (params of a
    primitive, threshold of a composite) — see strategies.resolve.
    """

    param_grid: dict[str, list[Any]]  # param key -> candidate values
    max_combos: int = 200  # safety cap on the grid size
    include_trades: bool = False  # include per-combo trade lists (heavy)
    sort_by: str = (
        "total_return_pct"  # total_return_pct|sharpe|profit_factor|max_drawdown_pct|num_trades
    )


class SweepResultItem(BaseModel):
    params: dict[str, Any]
    stats: StatsOut
    equity_curve: list[EquityPoint]
    trades: list[TradeOut] = []


class SweepResponse(BaseModel):
    strategy_slug: str
    exchange: str = "binance"
    symbols: list[str]
    interval: str
    start: int
    end: int
    count: int
    sort_by: str
    results: list[SweepResultItem]

from pydantic import BaseModel


class BacktestRequest(BaseModel):
    strategy_slug: str
    symbol: str
    interval: str
    start: str | int  # unix ms, RFC3339 or YYYY-MM-DD
    end: str | int
    initial_balance: str = "10000"
    fee_rate: str = "0.001"
    quantity: str | None = None  # fixed qty when the strategy doesn't dictate one


class TradeOut(BaseModel):
    time: int  # candle open time, ms
    side: str  # BUY / SELL
    price: str
    qty: str
    fee: str
    balance: str  # cash balance after the fill
    pnl: str | None = None  # realized pnl on sells


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
    symbol: str
    interval: str
    start: int
    end: int
    stats: StatsOut
    equity_curve: list[EquityPoint]
    trades: list[TradeOut]

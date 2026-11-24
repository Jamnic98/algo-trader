import math
from dataclasses import dataclass
from decimal import Decimal

from .binance import Candle
from .models import EquityPoint, StatsOut, TradeOut
from .strategies.base import CandleContext, Signal, Strategy

ZERO = Decimal(0)


class InsufficientFunds(Exception):
    pass


@dataclass
class PaperAccount:
    """Mirrors Go engine.PaperAccount (cash + long-only positions, per-fill fee)."""

    balance: Decimal
    fee_rate: Decimal
    positions: dict[str, Decimal]
    avg_entry: dict[str, Decimal]

    @classmethod
    def new(cls, balance: Decimal, fee_rate: Decimal) -> "PaperAccount":
        return cls(balance=balance, fee_rate=fee_rate, positions={}, avg_entry={})

    def apply(self, side: str, symbol: str, price: Decimal, qty: Decimal) -> Decimal:
        """Apply a fill, return the fee charged. Raises InsufficientFunds."""
        notional = price * qty
        fee = notional * self.fee_rate

        if side == "BUY":
            cost = notional + fee
            if self.balance < cost:
                raise InsufficientFunds("not enough cash")
            old_qty = self.positions.get(symbol, ZERO)
            total_qty = old_qty + qty
            if total_qty > 0:
                self.avg_entry[symbol] = (
                    self.avg_entry.get(symbol, ZERO) * old_qty + price * qty
                ) / total_qty
            self.balance -= cost
            self.positions[symbol] = total_qty
        elif side == "SELL":
            if self.positions.get(symbol, ZERO) < qty:
                raise InsufficientFunds("not enough asset")
            self.positions[symbol] -= qty
            if self.positions[symbol] == 0:
                self.avg_entry.pop(symbol, None)
            self.balance += notional - fee
        else:
            raise ValueError(f"invalid side {side!r}")

        return fee

    def equity(self, symbol: str, price: Decimal) -> Decimal:
        return self.balance + self.positions.get(symbol, ZERO) * price


def run_backtest(
    candles: list[Candle],
    strategy: Strategy,
    symbol: str,
    initial_balance: Decimal,
    fee_rate: Decimal,
    quantity: str | None = None,
) -> dict:
    """Replay candles through the strategy against a paper account.

    Buy sizing: strategy-decided qty, else the request's fixed `quantity`,
    else all-in (balance / (price * (1 + fee))). Sell sizing: strategy-decided
    qty, else the full open position.
    """
    account = PaperAccount.new(initial_balance, fee_rate)
    fixed_qty = Decimal(quantity) if quantity else None

    trades: list[TradeOut] = []
    equity_curve: list[EquityPoint] = []
    realized_pnl: list[Decimal] = []

    for c in candles:
        ctx = CandleContext(
            candle=c,
            fee_rate=fee_rate,
            avg_entry=account.avg_entry.get(symbol, ZERO),
        )
        decision = strategy.on_candle(ctx)

        if decision.signal == Signal.BUY:
            price = c.close
            qty = decision.quantity or fixed_qty
            if qty is None:
                qty = account.balance / (price * (ONE + fee_rate))
            if qty > 0:
                try:
                    fee = account.apply("BUY", symbol, price, qty)
                except InsufficientFunds:
                    fee = None
                if fee is not None:
                    trades.append(
                        TradeOut(
                            time=c.open_time,
                            side="BUY",
                            price=str(price),
                            qty=str(qty),
                            fee=str(fee),
                            balance=str(account.balance),
                        )
                    )
        elif decision.signal == Signal.SELL:
            qty = decision.quantity or account.positions.get(symbol, ZERO)
            if qty > 0:
                price = c.close
                entry = account.avg_entry.get(symbol, price)
                fee = price * qty * fee_rate
                pnl = (price - entry) * qty - fee
                try:
                    account.apply("SELL", symbol, price, qty)
                except InsufficientFunds:
                    pass
                else:
                    realized_pnl.append(pnl)
                    trades.append(
                        TradeOut(
                            time=c.open_time,
                            side="SELL",
                            price=str(price),
                            qty=str(qty),
                            fee=str(fee),
                            balance=str(account.balance),
                            pnl=str(pnl),
                        )
                    )

        equity_curve.append(
            EquityPoint(time=c.open_time, equity=str(account.equity(symbol, c.close)))
        )

    stats = _compute_stats(equity_curve, initial_balance, realized_pnl, len(trades))
    return {"stats": stats, "equity_curve": equity_curve, "trades": trades}


ONE = Decimal(1)


def _compute_stats(
    equity_curve: list[EquityPoint],
    initial_balance: Decimal,
    realized_pnl: list[Decimal],
    num_trades: int,
) -> StatsOut:
    equities = [Decimal(p.equity) for p in equity_curve]
    final = equities[-1] if equities else initial_balance

    total_return_pct = (
        ZERO
        if initial_balance == 0
        else (final - initial_balance) / initial_balance * 100
    )

    peak = ZERO
    max_dd_pct = ZERO
    for eq in equities:
        peak = max(peak, eq)
        if peak > 0:
            dd = (peak - eq) / peak * 100
            max_dd_pct = max(max_dd_pct, dd)

    wins = sum(1 for p in realized_pnl if p > 0)
    win_rate = (
        (Decimal(wins) / Decimal(len(realized_pnl)) * 100) if realized_pnl else ZERO
    )

    gross_profit = sum((p for p in realized_pnl if p > 0), ZERO)
    gross_loss = sum((-p for p in realized_pnl if p < 0), ZERO)
    if gross_loss == 0:
        profit_factor = None if gross_profit == 0 else "inf"
    else:
        profit_factor = str(gross_profit / gross_loss)

    sharpe = _sharpe(equities)

    return StatsOut(
        initial_balance=str(initial_balance),
        final_balance=str(final),
        total_return_pct=str(total_return_pct),
        max_drawdown_pct=str(max_dd_pct),
        num_trades=num_trades,
        win_rate=str(win_rate),
        profit_factor=profit_factor,
        sharpe=sharpe,
    )


def _sharpe(equities: list[Decimal]) -> str | None:
    if len(equities) < 2:
        return None
    returns = [
        float(equities[i] / equities[i - 1] - 1)
        for i in range(1, len(equities))
        if equities[i - 1] != 0
    ]
    if len(returns) < 2:
        return None
    mean = sum(returns) / len(returns)
    var = sum((r - mean) ** 2 for r in returns) / (len(returns) - 1)
    if var == 0:
        return None
    return str((mean / math.sqrt(var)) * math.sqrt(len(returns)))

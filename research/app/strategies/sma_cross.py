from decimal import Decimal

from .base import CandleContext, Decision, Signal


class SMACrossStrategy:
    """Mirrors Go strategies.SMACrossStrategy ("sma_cross" template): buy when the
    fast SMA crosses above the slow SMA, sell when it crosses below."""

    def __init__(self, params: dict) -> None:
        self.fast = int(params.get("fast") or 5)
        self.slow = int(params.get("slow") or 10)
        if self.fast < 1:
            self.fast = 5
        if self.slow < 2:
            self.slow = 10
        if self.fast >= self.slow:
            self.fast = self.slow - 1
        self.closes: list[Decimal] = []
        self.has_position = False

    def on_candle(self, ctx: CandleContext) -> Decision:
        self.closes.append(ctx.candle.close)
        if len(self.closes) > self.slow:
            self.closes = self.closes[-self.slow :]
        if len(self.closes) < self.slow:
            return Decision(Signal.HOLD)

        fast = sum(self.closes[-self.fast :], Decimal(0)) / self.fast
        slow = sum(self.closes, Decimal(0)) / self.slow

        if not self.has_position and fast > slow:
            self.has_position = True
            return Decision(Signal.BUY)
        if self.has_position and fast < slow:
            self.has_position = False
            return Decision(Signal.SELL)
        return Decision(Signal.HOLD)

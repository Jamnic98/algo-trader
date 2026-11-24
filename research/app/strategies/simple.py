from decimal import Decimal

from ..binance import Candle
from .base import CandleContext, Decision, Signal


class SimpleStrategy:
    """Mirrors Go strategies.SimpleStrategy: buy on a down candle, sell on an up
    candle, gated by a round-trip fee check."""

    def __init__(self, params: dict) -> None:
        self.has_position = False
        self.prev: Candle | None = None
        self.buy_price = Decimal(0)

    def on_candle(self, ctx: CandleContext) -> Decision:
        c = ctx.candle
        prev = self.prev
        self.prev = c

        if prev is None:
            return Decision(Signal.HOLD)

        if not self.has_position and c.close < prev.close:
            self.has_position = True
            self.buy_price = c.close
            return Decision(Signal.BUY)

        if self.has_position and c.close > prev.close:
            round_trip_fee = self.buy_price * ctx.fee_rate * 2
            price_diff = c.close - self.buy_price
            if price_diff <= round_trip_fee:
                return Decision(
                    Signal.HOLD,
                    reason=f"fee gate: move {price_diff} <= required {round_trip_fee}",
                )
            self.has_position = False
            self.buy_price = Decimal(0)
            return Decision(Signal.SELL)

        return Decision(Signal.HOLD)

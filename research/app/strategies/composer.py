from .base import CandleContext, Decision, Signal, Strategy


class WeightedComposer:
    """Mirrors Go strategies.WeightedComposer: weighted vote of sub-strategies,
    trade when the winning side's share of total weight reaches the threshold."""

    def __init__(self, threshold: float, nodes: list[tuple[Strategy, float]]) -> None:
        self.threshold = threshold
        self.nodes = nodes

    def on_candle(self, ctx: CandleContext) -> Decision:
        buy_score = 0.0
        sell_score = 0.0
        total_weight = 0.0

        for strategy, weight in self.nodes:
            d = strategy.on_candle(ctx)
            total_weight += weight
            if d.signal == Signal.BUY:
                buy_score += weight
            elif d.signal == Signal.SELL:
                sell_score += weight

        if total_weight == 0:
            return Decision(Signal.HOLD)

        buy_score /= total_weight
        sell_score /= total_weight

        if buy_score >= self.threshold:
            return Decision(Signal.BUY)
        if sell_score >= self.threshold:
            return Decision(Signal.SELL)
        return Decision(
            Signal.HOLD,
            reason=f"buy={buy_score:.2f} sell={sell_score:.2f} threshold={self.threshold:.2f}",
        )

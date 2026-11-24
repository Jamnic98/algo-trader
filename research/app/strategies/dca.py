from decimal import Decimal

from .base import CandleContext, Decision, Signal


class SimpleDCAStrategy:
    """Mirrors Go strategies.SimpleDCAStrategy: buys every candle."""

    def __init__(self, params: dict) -> None:
        pass

    def on_candle(self, ctx: CandleContext) -> Decision:
        return Decision(Signal.BUY)


class SmartDCAStrategy:
    """Mirrors Go strategies.SmartDCAStrategy: fixed-spend DCA buys, sells once
    price exceeds avg entry * take_profit and the sell fee is covered."""

    def __init__(self, params: dict) -> None:
        fixed_spend = Decimal(str(params.get("fixed_spend") or "10.00"))
        if fixed_spend <= 0:
            fixed_spend = Decimal(10)
        take_profit = Decimal(str(params.get("take_profit") or "1.05"))
        if take_profit <= 0:
            take_profit = Decimal("1.05")
        self.fixed_spend = fixed_spend
        self.take_profit = take_profit
        self.total_spent = Decimal(0)
        self.total_bought = Decimal(0)

    def on_candle(self, ctx: CandleContext) -> Decision:
        price = ctx.candle.close

        if self.total_bought > 0:
            avg_entry = self.total_spent / self.total_bought
            threshold = avg_entry * self.take_profit

            if price > threshold:
                sell_value = price * self.total_bought
                sell_fee = sell_value * ctx.fee_rate
                profit = sell_value - self.total_spent - sell_fee
                if profit < 0:
                    return Decision(Signal.HOLD)

                qty_to_sell = self.total_bought
                self.total_spent = Decimal(0)
                self.total_bought = Decimal(0)
                return Decision(Signal.SELL, quantity=qty_to_sell)

        qty_to_buy = self.fixed_spend / price
        cost = price * qty_to_buy
        buy_fee = cost * ctx.fee_rate

        if buy_fee >= self.fixed_spend:
            return Decision(Signal.HOLD)

        self.total_spent += cost + buy_fee  # include fee in cost basis
        self.total_bought += qty_to_buy
        return Decision(Signal.BUY, quantity=qty_to_buy)

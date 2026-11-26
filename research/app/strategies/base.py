from dataclasses import dataclass, field
from decimal import Decimal
from enum import Enum
from typing import Protocol

from ..binance import Candle


class Signal(str, Enum):
    HOLD = "HOLD"
    BUY = "BUY"
    SELL = "SELL"


@dataclass
class Decision:
    signal: Signal
    quantity: Decimal | None = None  # None = caller decides
    reason: str = ""


@dataclass
class CandleContext:
    candle: Candle
    fee_rate: Decimal
    avg_entry: Decimal = field(default_factory=lambda: Decimal(0))  # 0 if no position
    symbol: str = ""  # which symbol this candle belongs to (multi-symbol runs)


class Strategy(Protocol):
    def on_candle(self, ctx: CandleContext) -> Decision: ...

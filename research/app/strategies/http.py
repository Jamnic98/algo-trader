from decimal import Decimal
from typing import Any

import httpx

from .base import CandleContext, Decision, Signal

VALID_SIGNALS = {s.value for s in Signal}


class HTTPStrategy:
    """Mirrors Go strategies.HTTPStrategy ("http" template): strategy logic lives
    in an external service (e.g. an ML inference endpoint). Each candle, POST the
    recent close window plus params; expect {"signal": ..., "quantity": ...}."""

    def __init__(
        self, params: dict[str, Any], client: httpx.Client | None = None
    ) -> None:
        url = params.get("url")
        if not url:
            raise ValueError('http strategy requires a "url" param')
        self.url: str = str(url)
        self.window = int(params.get("window") or 100)
        if self.window <= 0:
            self.window = 100
        self.extra = {k: v for k, v in params.items() if k not in ("url", "window")}
        self.closes: list[str] = []
        self.client = client or httpx.Client(timeout=5.0)

    def on_candle(self, ctx: CandleContext) -> Decision:
        self.closes.append(str(ctx.candle.close))
        if len(self.closes) > self.window:
            self.closes = self.closes[-self.window :]

        try:
            resp = self.client.post(
                self.url, json={"closes": self.closes, "params": self.extra}
            )
            resp.raise_for_status()
            data = resp.json()
        except Exception as exc:  # unreachable / bad response -> safe HOLD
            return Decision(Signal.HOLD, reason=f"http strategy error: {exc}")

        sig = str(data.get("signal", "HOLD")).upper()
        if sig not in VALID_SIGNALS:
            return Decision(Signal.HOLD, reason=f"invalid signal {sig!r}")

        raw_qty = data.get("quantity")
        qty = Decimal(str(raw_qty)) if raw_qty not in (None, "") else None
        return Decision(Signal(sig), quantity=qty)

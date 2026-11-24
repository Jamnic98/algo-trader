import json
from decimal import Decimal

import httpx

from app.binance import Candle
from app.strategies import build_primitive
from app.strategies.base import CandleContext, Signal
from app.strategies.http import HTTPStrategy


def candle(close: str) -> Candle:
    price = Decimal(close)
    return Candle(
        open_time=0,
        close_time=60_000,
        open=price,
        high=price,
        low=price,
        close=price,
        volume=Decimal(1),
    )


def ctx(c: Candle) -> CandleContext:
    return CandleContext(candle=c, fee_rate=Decimal("0.001"))


def make_client(handler) -> httpx.Client:
    return httpx.Client(transport=httpx.MockTransport(handler), base_url="http://ml")


def test_http_strategy_signals_window_and_params() -> None:
    seen = {}

    def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content)
        seen.update(body)
        closes = [float(c) for c in body["closes"]]
        if len(closes) >= 2 and closes[-1] > closes[-2]:
            return httpx.Response(200, json={"signal": "BUY", "quantity": "0.5"})
        if len(closes) >= 2 and closes[-1] < closes[-2]:
            return httpx.Response(200, json={"signal": "SELL"})
        return httpx.Response(200, json={"signal": "HOLD"})

    strat = HTTPStrategy(
        {"url": "/predict", "window": 3, "model": "toy"}, client=make_client(handler)
    )

    assert strat.on_candle(ctx(candle("100"))).signal == Signal.HOLD
    d = strat.on_candle(ctx(candle("101")))
    assert d.signal == Signal.BUY and d.quantity == Decimal("0.5")

    strat.on_candle(ctx(candle("102")))
    assert len(seen["closes"]) == 3  # window capped
    assert seen["params"] == {"model": "toy"}  # url/window stripped

    assert strat.on_candle(ctx(candle("90"))).signal == Signal.SELL


def test_http_strategy_failures_hold() -> None:
    def boom(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500)

    strat = HTTPStrategy({"url": "/predict"}, client=make_client(boom))
    d = strat.on_candle(ctx(candle("100")))
    assert d.signal == Signal.HOLD and "error" in d.reason

    def bad_signal(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"signal": "YOLO"})

    strat2 = HTTPStrategy({"url": "/predict"}, client=make_client(bad_signal))
    assert strat2.on_candle(ctx(candle("100"))).signal == Signal.HOLD


def test_http_strategy_requires_url() -> None:
    import pytest

    with pytest.raises(ValueError, match="url"):
        HTTPStrategy({})


def test_ml_example_registered_with_schema_defaults() -> None:
    strat = build_primitive("ml_example", {"url": "http://127.0.0.1:1/predict"})
    assert isinstance(strat, HTTPStrategy)
    assert strat.window == 20  # schema default
    assert strat.url == "http://127.0.0.1:1/predict"

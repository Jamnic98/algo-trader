"""Runs the shared behavior vectors from shared/strategies.json against the
Python implementations. The Go suite runs the same vectors
(trader-core/internal/strategies/definitions_test.go) — if either fails,
the engines have drifted from the shared schema."""

from decimal import Decimal
from typing import cast

from app.binance import Candle
from app.strategies import build_primitive
from app.strategies.base import CandleContext
from app.strategies.dca import SmartDCAStrategy
from app.strategies.definitions import (
    composite_default_threshold,
    find_primitive,
    load_schema,
)


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


def test_schema_loads_and_has_primitives() -> None:
    schema = load_schema()
    names = [d["name"] for d in schema["primitives"]]
    assert {"simple", "simple_dca", "smart_dca"} <= set(names)
    assert schema["composite"]["template"] == "weighted_vote"
    assert composite_default_threshold() == 0.6


def test_every_primitive_is_buildable() -> None:
    for defn in load_schema()["primitives"]:
        assert find_primitive(defn["name"]) is not None
        build_primitive(defn["name"], {})


def test_shared_test_vectors() -> None:
    for defn in load_schema()["primitives"]:
        fee = Decimal(defn["test"]["fee_rate"])
        for case in defn["test"]["cases"]:
            strat = build_primitive(defn["name"], {})
            signals = []
            for close in case["candles"]:
                d = strat.on_candle(
                    CandleContext(candle=candle(str(close)), fee_rate=fee)
                )
                signals.append(d.signal.value)
            assert (
                signals == case["expect"]
            ), f"{defn['name']}: {signals} != {case['expect']}"


def test_param_defaults_come_from_schema() -> None:
    strat = cast(SmartDCAStrategy, build_primitive("smart_dca", {}))
    assert strat.fixed_spend == Decimal("10.00")
    assert strat.take_profit == Decimal("1.05")

    overridden = cast(
        SmartDCAStrategy,
        build_primitive("smart_dca", {"fixed_spend": "25.00", "take_profit": 1.1}),
    )
    assert overridden.fixed_spend == Decimal("25.00")
    assert overridden.take_profit == Decimal("1.1")


def test_unknown_strategy_and_template() -> None:
    import pytest

    with pytest.raises(ValueError, match="unknown strategy"):
        build_primitive("nope", {})

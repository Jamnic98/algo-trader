from decimal import Decimal

import pytest

from app.binance import Candle
from app.strategies import build_primitive, resolve
from app.strategies.base import CandleContext, Signal
from app.strategies.simple import SimpleStrategy

FEE = Decimal("0.001")


def candle(close: str, open_time: int = 0) -> Candle:
    price = Decimal(close)
    return Candle(
        open_time=open_time,
        close_time=open_time + 60_000,
        open=price,
        high=price,
        low=price,
        close=price,
        volume=Decimal(1),
    )


def ctx(c: Candle) -> CandleContext:
    return CandleContext(candle=c, fee_rate=FEE, avg_entry=Decimal(0))


def test_simple_buys_on_down_candle_and_sells_on_up_candle() -> None:
    s = SimpleStrategy({})

    assert s.on_candle(ctx(candle("100"))).signal == Signal.HOLD  # no prev
    assert s.on_candle(ctx(candle("90"))).signal == Signal.BUY  # down -> buy
    assert (
        s.on_candle(ctx(candle("95"))).signal == Signal.SELL
    )  # up -> sell (move 5 > fee)
    assert s.on_candle(ctx(candle("94"))).signal == Signal.BUY  # down again -> buy


def test_simple_fee_gate_blocks_small_moves() -> None:
    s = SimpleStrategy({})
    s.on_candle(ctx(candle("100")))
    s.on_candle(ctx(candle("90")))  # buy @90
    # up move of 0.05 on price 90 is below round-trip fee 90*0.002 = 0.18
    d = s.on_candle(ctx(candle("90.05")))
    assert d.signal == Signal.HOLD
    assert "fee gate" in d.reason
    assert s.has_position  # still holding


def test_simple_dca_always_buys() -> None:
    s = build_primitive("simple_dca", {})
    for price in ("100", "90", "95"):
        assert s.on_candle(ctx(candle(price))).signal == Signal.BUY


def test_smart_dca_buys_fixed_spend_and_sells_above_take_profit() -> None:
    s = build_primitive("smart_dca", {"fixed_spend": "10.00"})

    d = s.on_candle(ctx(candle("100")))
    assert d.signal == Signal.BUY
    assert d.quantity == Decimal("0.1")  # 10 / 100

    # price 106 > avg entry 100 * 1.05 -> sell everything
    d = s.on_candle(ctx(candle("106")))
    assert d.signal == Signal.SELL
    assert d.quantity == Decimal("0.1")


def test_smart_dca_holds_below_take_profit() -> None:
    s = build_primitive("smart_dca", {"fixed_spend": "10.00"})
    s.on_candle(ctx(candle("100")))
    d = s.on_candle(ctx(candle("102")))  # below 105 threshold
    assert d.signal == Signal.BUY  # keeps DCA-ing


def test_resolve_composite_builds_weighted_composer(store) -> None:
    row = store.get("combo")
    strat = resolve(row, None, store)

    # first candle: simple holds (no prev), simple_dca buys -> buy score 0.5 < 0.6 -> HOLD
    assert strat.on_candle(ctx(candle("100"))).signal == Signal.HOLD
    # second candle down: simple buys, simple_dca buys -> buy score 1.0 >= 0.6 -> BUY
    assert strat.on_candle(ctx(candle("90"))).signal == Signal.BUY


def test_resolve_unknown_strategy_raises(store) -> None:
    from app.store import StrategyRow

    row = StrategyRow(
        slug="nope", display_name="Nope", is_composite=False, default_config={}
    )
    with pytest.raises(ValueError, match="unknown strategy"):
        resolve(row, None, store)


def test_resolve_composite_missing_nodes_raises(store) -> None:
    from app.store import StrategyRow

    row = StrategyRow(
        slug="empty",
        display_name="Empty",
        is_composite=True,
        default_config={},
        nodes=[],
    )
    with pytest.raises(ValueError, match="no nodes"):
        resolve(row, None, store)


def test_resolve_nested_composite(store) -> None:
    # combo_nested = weighted_vote(combo(simple_dca + simple), simple_dca)
    strat = resolve(store.get("combo_nested"), None, store)

    # c1: inner combo 1/2 -> HOLD, dca BUY -> 1/2 votes -> HOLD
    assert strat.on_candle(ctx(candle("100"))).signal == Signal.HOLD
    # c2 (down): inner combo BUY (2/2), dca BUY -> 2/2 -> BUY
    assert strat.on_candle(ctx(candle("90"))).signal == Signal.BUY

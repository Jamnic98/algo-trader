from decimal import Decimal

import pytest

from app.binance import Candle
from app.engine import PaperAccount, InsufficientFunds, _compute_stats, run_backtest
from app.models import EquityPoint
from app.strategies import build_primitive

FEE = Decimal("0.001")


def make_candles(
    closes: list[str], start: int = 1_700_000_000_000, step: int = 60_000
) -> list[Candle]:
    out = []
    for i, close in enumerate(closes):
        price = Decimal(close)
        out.append(
            Candle(
                open_time=start + i * step,
                close_time=start + i * step + step - 1,
                open=price,
                high=price,
                low=price,
                close=price,
                volume=Decimal(1),
            )
        )
    return out


def test_paper_account_buy_sell_roundtrip() -> None:
    acc = PaperAccount.new(Decimal("1000"), FEE)
    fee = acc.apply("BUY", "BTCUSDT", Decimal("100"), Decimal("2"))
    assert fee == Decimal("0.2")
    assert acc.balance == Decimal("1000") - Decimal("200") - Decimal("0.2")
    assert acc.positions["BTCUSDT"] == Decimal("2")
    assert acc.avg_entry["BTCUSDT"] == Decimal("100")

    acc.apply("SELL", "BTCUSDT", Decimal("110"), Decimal("2"))
    assert acc.positions["BTCUSDT"] == Decimal("0")
    assert "BTCUSDT" not in acc.avg_entry
    # 1000 - 200.2 + 220 - 0.22 = 1019.58
    assert acc.balance == Decimal("1019.58")


def test_paper_account_rejects_overspend() -> None:
    acc = PaperAccount.new(Decimal("100"), FEE)
    try:
        acc.apply("BUY", "BTCUSDT", Decimal("100"), Decimal("2"))
        raise AssertionError("expected InsufficientFunds")
    except InsufficientFunds:
        pass
    try:
        acc.apply("SELL", "BTCUSDT", Decimal("100"), Decimal("1"))
        raise AssertionError("expected InsufficientFunds")
    except InsufficientFunds:
        pass


def test_backtest_simple_dca_fixed_quantity() -> None:
    candles = make_candles(["100", "100", "100"])
    strat = build_primitive("simple_dca", {})
    result = run_backtest(candles, strat, "BTCUSDT", Decimal("1000"), FEE, quantity="5")

    # first buy costs 5*100*1.001 = 500.5, second would cost the same -> only 1 trade
    assert result["stats"].num_trades == 1
    assert result["trades"][0].side == "BUY"
    assert result["trades"][0].qty == "5"
    assert len(result["equity_curve"]) == 3
    # price flat -> final_balance (equity incl. holdings) = initial minus fee
    final = Decimal(result["stats"].final_balance)
    assert final < Decimal("1000")
    assert final > Decimal("998")


def test_backtest_simple_dca_all_in_when_no_quantity() -> None:
    candles = make_candles(["100", "100"])
    strat = build_primitive("simple_dca", {})
    result = run_backtest(candles, strat, "BTCUSDT", Decimal("1000"), FEE)

    assert result["trades"][0].side == "BUY"
    # all-in sizing: qty * price * (1 + fee) ~= initial balance
    qty = Decimal(result["trades"][0].qty)
    assert float(qty) * 100 * 1.001 == pytest.approx(1000)
    final = Decimal(result["stats"].final_balance)
    assert final < Decimal("1000")
    assert final > Decimal("998")


def test_backtest_simple_strategy_round_trip_profit() -> None:
    # down -> buy @90, up -> sell @100 (move covers fees)
    candles = make_candles(["100", "90", "100"])
    strat = build_primitive("simple", {})
    result = run_backtest(candles, strat, "BTCUSDT", Decimal("1000"), FEE)

    sides = [t.side for t in result["trades"]]
    assert sides == ["BUY", "SELL"]
    sell = result["trades"][1]
    assert sell.pnl is not None and Decimal(sell.pnl) > 0
    assert Decimal(result["stats"].final_balance) > Decimal("1000")
    assert Decimal(result["stats"].total_return_pct) > 0
    assert result["stats"].win_rate == "100"


def test_compute_stats_max_drawdown_and_sharpe() -> None:
    equity = [
        EquityPoint(time=i, equity=e) for i, e in enumerate(["100", "120", "90", "110"])
    ]
    stats = _compute_stats(
        equity, Decimal("100"), [Decimal("5"), Decimal("-2")], num_trades=2
    )

    assert Decimal(stats.total_return_pct) == Decimal("10")  # 110 vs 100
    # peak 120 -> trough 90 -> dd = 25%
    assert Decimal(stats.max_drawdown_pct) == Decimal("25")
    assert Decimal(stats.win_rate) == Decimal("50")
    assert stats.profit_factor is not None
    assert Decimal(stats.profit_factor) == Decimal("2.5")  # 5 / 2
    assert stats.sharpe is not None


def test_compute_stats_profit_factor_no_losses() -> None:
    equity = [EquityPoint(time=0, equity="100"), EquityPoint(time=1, equity="110")]
    stats = _compute_stats(equity, Decimal("100"), [Decimal("10")], num_trades=1)
    assert stats.profit_factor == "inf"


def test_backtest_records_equity_even_when_fill_fails() -> None:
    # regression: a rejected fill must not skip the equity point for that candle
    candles = make_candles(["100", "100", "100"])
    strat = build_primitive("simple_dca", {})
    result = run_backtest(candles, strat, "BTCUSDT", Decimal("100"), FEE, quantity="5")

    # first buy costs 500.5 > 100 balance -> rejected; equity still recorded per candle
    assert result["stats"].num_trades == 0
    assert len(result["equity_curve"]) == 3
    assert all(Decimal(p.equity) == Decimal("100") for p in result["equity_curve"])


def test_backtest_sell_without_position_is_skipped() -> None:
    class AlwaysSell:
        def on_candle(self, ctx):
            from app.strategies.base import Decision, Signal

            return Decision(Signal.SELL)

    candles = make_candles(["100", "110"])
    result = run_backtest(candles, AlwaysSell(), "BTCUSDT", Decimal("1000"), FEE)

    assert result["stats"].num_trades == 0
    assert len(result["equity_curve"]) == 2
    assert Decimal(result["stats"].final_balance) == Decimal("1000")


def test_backtest_fixed_quantity_sell_closes_position() -> None:
    # simple: buy @90 then sell @100 with fixed qty 1
    candles = make_candles(["100", "90", "100"])
    strat = build_primitive("simple", {})
    result = run_backtest(candles, strat, "BTCUSDT", Decimal("1000"), FEE, quantity="1")

    sides = [t.side for t in result["trades"]]
    assert sides == ["BUY", "SELL"]
    assert all(t.qty == "1" for t in result["trades"])
    # pnl = (100-90)*1 - sell fee 0.1 = 9.9
    assert Decimal(result["trades"][1].pnl) == Decimal("9.9")


def test_backtest_slippage_worsens_fills() -> None:
    closes = ["100", "90", "100", "110"]
    base = run_backtest(
        make_candles(closes),
        build_primitive("simple", {}),
        "BTCUSDT",
        Decimal(1000),
        FEE,
    )
    slip = run_backtest(
        make_candles(closes),
        build_primitive("simple", {}),
        "BTCUSDT",
        Decimal(1000),
        FEE,
        slippage_bps=Decimal(50),  # 0.5%
    )
    # no slippage: fills exactly at close
    assert Decimal(base["trades"][0].price) == Decimal("90")
    assert Decimal(base["trades"][1].price) == Decimal("100")
    # with slippage: buy pays more, sell receives less
    assert Decimal(slip["trades"][0].price) == Decimal("90") * Decimal("1.005")
    assert Decimal(slip["trades"][1].price) == Decimal("100") * Decimal("0.995")
    # and the run performs worse
    assert Decimal(slip["stats"].final_balance) < Decimal(base["stats"].final_balance)


def test_run_portfolio_multi_symbol_shared_account() -> None:
    from app.engine import run_portfolio

    start = 1_700_000_000_000
    btc = make_candles(["100", "90", "100"], start=start)
    eth = make_candles(["10", "9", "10"], start=start)
    # one strategy instance per symbol (strategies keep internal state)
    strats = {s: build_primitive("simple", {}) for s in ("BTCUSDT", "ETHUSDT")}
    res = run_portfolio({"BTCUSDT": btc, "ETHUSDT": eth}, strats, Decimal(1000), FEE)

    assert res["symbols"] == ["BTCUSDT", "ETHUSDT"]
    buys = [t for t in res["trades"] if t.side == "BUY"]
    sells = [t for t in res["trades"] if t.side == "SELL"]
    # both symbols traded independently despite sharing the cash account
    assert {t.symbol for t in buys} == {"BTCUSDT", "ETHUSDT"}
    assert {t.symbol for t in sells} == {"BTCUSDT", "ETHUSDT"}
    # one equity point per timestamp, not per symbol-candle
    assert len(res["equity_curve"]) == 3
    # all-in sizing splits the balance across the 2 symbols, and the run profits
    assert Decimal(res["stats"].final_balance) > Decimal(1000)


def test_run_portfolio_requires_per_symbol_instances() -> None:
    # strategies keep internal state, so sharing one instance across symbols
    # would corrupt signals — the engine requires a per-symbol mapping and
    # run_backtest is the single-symbol convenience wrapper.
    from app.engine import run_portfolio

    start = 1_700_000_000_000
    candles = make_candles(["100", "90", "100"], start=start)
    res = run_portfolio(
        {"AAA": candles, "BBB": make_candles(["10", "9", "10"], start=start)},
        {"AAA": build_primitive("simple", {}), "BBB": build_primitive("simple", {})},
        Decimal(1000),
        FEE,
    )
    buys = [t for t in res["trades"] if t.side == "BUY"]
    assert {t.symbol for t in buys} == {"AAA", "BBB"}


def test_run_portfolio_multi_symbol_slippage() -> None:
    from app.engine import run_portfolio

    start = 1_700_000_000_000
    candles = make_candles(["100", "90", "100"], start=start)
    strats = {"AAA": build_primitive("simple", {})}
    res = run_portfolio(
        {"AAA": candles}, strats, Decimal(1000), FEE, slippage_bps=Decimal(100)
    )
    # 1% slippage: buy at 90 * 1.01
    buy = next(t for t in res["trades"] if t.side == "BUY")
    assert Decimal(buy.price) == Decimal("90") * Decimal("1.01")

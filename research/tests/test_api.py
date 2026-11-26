from collections.abc import Iterator

import httpx
import pytest
from fastapi.testclient import TestClient

from app.main import app, get_http_client, get_store


def kline_row(open_time: int, close: str) -> list:
    return [
        open_time,
        close,
        close,
        close,
        close,
        "1.0",
        open_time + 3_599_999,
        "0",
        0,
        "0",
        "0",
        "0",
    ]


@pytest.fixture
def client(store) -> Iterator[TestClient]:
    def fake_klines(request: httpx.Request) -> httpx.Response:
        start = int(request.url.params["startTime"])
        closes = ["100", "90", "100", "110"]
        rows = [kline_row(start + i * 3_600_000, c) for i, c in enumerate(closes)]
        return httpx.Response(200, json=rows)

    app.dependency_overrides[get_store] = lambda: store
    app.dependency_overrides[get_http_client] = lambda: httpx.AsyncClient(
        transport=httpx.MockTransport(fake_klines), base_url="https://api.binance.com"
    )
    yield TestClient(app)
    app.dependency_overrides.clear()


def test_health(client: TestClient) -> None:
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_backtest_ok(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": "2024-01-01",
            "end": 1704067200000 + 4 * 3_600_000,
            "initial_balance": "1000",
            "fee_rate": "0.001",
        },
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["strategy_slug"] == "simple"
    assert body["symbol"] == "BTCUSDT"
    assert body["stats"]["num_trades"] == 2  # buy @90, sell @100
    assert [t["side"] for t in body["trades"]] == ["BUY", "SELL"]
    assert len(body["equity_curve"]) == 4
    assert body["stats"]["final_balance"] > body["stats"]["initial_balance"]


def test_backtest_composite_strategy(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "combo",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704067200000 + 4 * 3_600_000,
        },
    )
    assert resp.status_code == 200, resp.text
    assert resp.json()["strategy_slug"] == "combo"


def test_backtest_unknown_strategy_404(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "nope",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
        },
    )
    assert resp.status_code == 404


def test_backtest_bad_interval_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "7m",
            "start": 1704067200000,
            "end": 1704070800000,
        },
    )
    assert resp.status_code == 400


def test_backtest_end_before_start_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704070800000,
            "end": 1704067200000,
        },
    )
    assert resp.status_code == 400


def test_backtest_invalid_decimal_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
            "initial_balance": "not-a-number",
        },
    )
    assert resp.status_code == 400


def test_backtest_no_klines_404(client: TestClient) -> None:
    import httpx

    from app.main import get_http_client

    def empty_klines(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json=[])

    app.dependency_overrides[get_http_client] = lambda: httpx.AsyncClient(
        transport=httpx.MockTransport(empty_klines), base_url="https://api.binance.com"
    )
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
        },
    )
    assert resp.status_code == 404
    assert "no klines" in resp.json()["detail"]


def test_backtest_binance_error_502(client: TestClient) -> None:
    import httpx

    from app.main import get_http_client

    def boom(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500, text="server error")

    app.dependency_overrides[get_http_client] = lambda: httpx.AsyncClient(
        transport=httpx.MockTransport(boom), base_url="https://api.binance.com"
    )
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
        },
    )
    assert resp.status_code == 502


@pytest.fixture
def alpaca_client(store) -> Iterator[TestClient]:
    from app.main import get_alpaca_client

    def fake_bars(request: httpx.Request) -> httpx.Response:
        assert (
            request.url.raw_path.split(b"?")[0] == b"/v1beta3/crypto/us/BTC%2FUSD/bars"
        )
        closes = ["50000", "49000", "51000", "52000"]
        bars = [
            {
                "t": f"2024-01-01T0{i}:00:00Z",
                "o": c,
                "h": c,
                "l": c,
                "c": c,
                "v": "1.5",
            }
            for i, c in enumerate(closes)
        ]
        return httpx.Response(200, json={"bars": bars, "next_page_token": ""})

    app.dependency_overrides[get_store] = lambda: store
    app.dependency_overrides[get_http_client] = lambda: httpx.AsyncClient(
        transport=httpx.MockTransport(lambda r: httpx.Response(200, json=[])),
        base_url="https://api.binance.com",
    )
    app.dependency_overrides[get_alpaca_client] = lambda: httpx.AsyncClient(
        transport=httpx.MockTransport(fake_bars),
        base_url="https://data.alpaca.markets",
        headers={"APCA-API-KEY-ID": "k", "APCA-API-SECRET-KEY": "s"},
    )
    yield TestClient(app)
    app.dependency_overrides.clear()


def test_backtest_alpaca_exchange(alpaca_client: TestClient) -> None:
    resp = alpaca_client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "exchange": "alpaca",
            "symbol": "btc/usd",
            "interval": "1h",
            "start": "2024-01-01",
            "end": 1704067200000 + 4 * 3_600_000,
            "initial_balance": "1000",
            "fee_rate": "0.001",
        },
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["exchange"] == "alpaca"
    assert body["symbol"] == "BTC/USD"  # normalized
    assert body["stats"]["num_trades"] >= 1


def test_backtest_unsupported_exchange_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "exchange": "kraken",
            "symbol": "BTC/USD",
            "interval": "1h",
            "start": "2024-01-01",
            "end": "2024-02-01",
        },
    )
    assert resp.status_code == 400
    assert "unsupported exchange" in resp.json()["detail"]


def test_backtest_alpaca_bad_interval_400(alpaca_client: TestClient) -> None:
    resp = alpaca_client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "exchange": "alpaca",
            "symbol": "AAPL",
            "interval": "3m",  # not offered by Alpaca
            "start": "2024-01-01",
            "end": "2024-02-01",
        },
    )
    assert resp.status_code == 400
    assert "unsupported interval" in resp.json()["detail"]


def test_backtest_multi_symbol_and_slippage(client: TestClient) -> None:
    from decimal import Decimal

    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "symbols": ["BTCUSDT", "ETHUSDT"],
            "interval": "1h",
            "start": "2024-01-01",
            "end": 1704067200000 + 4 * 3_600_000,
            "initial_balance": "1000",
            "fee_rate": "0.001",
            "slippage_bps": "50",
        },
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["symbols"] == ["BTCUSDT", "ETHUSDT"]
    assert body["symbol"] == "BTCUSDT"  # back-compat: first symbol
    buys = [t for t in body["trades"] if t["side"] == "BUY"]
    assert {t["symbol"] for t in buys} == {"BTCUSDT", "ETHUSDT"}
    # 0.5% slippage: the 90-close buy fills at 90 * 1.005 = 90.45
    assert all(Decimal(t["price"]) == Decimal("90") * Decimal("1.005") for t in buys)


def test_backtest_requires_a_symbol(client: TestClient) -> None:
    resp = client.post(
        "/backtest",
        json={
            "strategy_slug": "simple",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
        },
    )
    assert resp.status_code == 400


def test_backtest_sweep_ok(client: TestClient) -> None:
    from decimal import Decimal

    resp = client.post(
        "/backtest/sweep",
        json={
            "strategy_slug": "smart_dca",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": "2024-01-01",
            "end": 1704067200000 + 4 * 3_600_000,
            "initial_balance": "1000",
            "fee_rate": "0.001",
            "param_grid": {
                "fixed_spend": ["5", "10", "20"],
                "take_profit": [1.05, 1.1],
            },
            "sort_by": "total_return_pct",
        },
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["count"] == 6
    assert len(body["results"]) == 6
    assert body["symbols"] == ["BTCUSDT"]
    # every cartesian combo is present, echoing its params
    combos = {
        (r["params"]["fixed_spend"], r["params"]["take_profit"])
        for r in body["results"]
    }
    assert combos == {(f, tp) for f in ("5", "10", "20") for tp in (1.05, 1.1)}
    # sorted by total_return_pct descending
    rets = [Decimal(r["stats"]["total_return_pct"]) for r in body["results"]]
    assert rets == sorted(rets, reverse=True)
    # trades omitted unless requested
    assert all(r["trades"] == [] for r in body["results"])


def test_backtest_sweep_include_trades_and_max_drawdown_sort(
    client: TestClient,
) -> None:
    resp = client.post(
        "/backtest/sweep",
        json={
            "strategy_slug": "smart_dca",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": "2024-01-01",
            "end": 1704067200000 + 4 * 3_600_000,
            "param_grid": {"fixed_spend": ["5", "20"]},
            "include_trades": True,
            "sort_by": "max_drawdown_pct",  # ascending
        },
    )
    assert resp.status_code == 200, resp.text
    body = resp.json()
    assert body["count"] == 2
    assert all(len(r["trades"]) > 0 for r in body["results"])
    dds = [float(r["stats"]["max_drawdown_pct"]) for r in body["results"]]
    assert dds == sorted(dds)


def test_backtest_sweep_empty_grid_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest/sweep",
        json={
            "strategy_slug": "smart_dca",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
            "param_grid": {},
        },
    )
    assert resp.status_code == 400


def test_backtest_sweep_grid_too_large_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest/sweep",
        json={
            "strategy_slug": "smart_dca",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
            "param_grid": {
                "fixed_spend": [str(i) for i in range(10)],
                "take_profit": [1.05, 1.1, 1.2, 1.3],
            },
            "max_combos": 20,  # 40 combos > 20
        },
    )
    assert resp.status_code == 400
    assert "too large" in resp.json()["detail"]


def test_backtest_sweep_bad_sort_by_400(client: TestClient) -> None:
    resp = client.post(
        "/backtest/sweep",
        json={
            "strategy_slug": "smart_dca",
            "symbol": "BTCUSDT",
            "interval": "1h",
            "start": 1704067200000,
            "end": 1704070800000,
            "param_grid": {"fixed_spend": ["5"]},
            "sort_by": "nope",
        },
    )
    assert resp.status_code == 400

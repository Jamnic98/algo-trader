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

import httpx
import pytest

from app.alpaca import (
    bars_path,
    fetch_bars,
    is_crypto,
    normalize_symbol,
    parse_bar_time,
)


def make_client(handler) -> httpx.AsyncClient:
    return httpx.AsyncClient(
        transport=httpx.MockTransport(handler), base_url="https://data.alpaca.markets"
    )


def bar(t: str, c: str) -> dict:
    return {"t": t, "o": c, "h": c, "l": c, "c": c, "v": "10"}


def test_normalize_symbol() -> None:
    assert normalize_symbol("btc/usd") == "BTC/USD"
    assert normalize_symbol("aapl") == "AAPL"


def test_is_crypto() -> None:
    assert is_crypto("BTC/USD")
    assert not is_crypto("AAPL")


def test_bars_path() -> None:
    assert bars_path("AAPL") == "/v2/stocks/AAPL/bars"
    assert bars_path("BTC/USD") == "/v1beta3/crypto/us/BTC%2FUSD/bars"


def test_parse_bar_time() -> None:
    assert parse_bar_time("2024-01-01T09:30:00Z") == 1704101400000
    assert parse_bar_time("2024-01-01T09:30:00.123456789Z") == 1704101400000
    assert parse_bar_time("2024-01-01T09:30:00+00:00") == 1704101400000
    with pytest.raises(ValueError):
        parse_bar_time("not-a-time")


async def test_fetch_bars_stocks_paginates() -> None:
    calls: list[dict] = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(dict(request.url.params))
        assert request.url.path == "/v2/stocks/AAPL/bars"
        if len(calls) == 1:
            assert calls[0]["feed"] == "iex"
            assert calls[0]["timeframe"] == "1Min"
            assert "page_token" not in calls[0]
            return httpx.Response(
                200,
                json={
                    "bars": [bar("2024-01-01T09:30:00Z", "100")],
                    "next_page_token": "page2",
                },
            )
        assert calls[1]["page_token"] == "page2"
        return httpx.Response(
            200,
            json={
                "bars": [bar("2024-01-01T09:31:00Z", "101")],
                "next_page_token": None,
            },
        )

    candles = await fetch_bars(
        make_client(handler), "aapl", "1m", 1704101400000, 1704101400000 + 3_600_000
    )
    assert len(calls) == 2
    assert len(candles) == 2
    assert candles[0].open_time == 1704101400000
    assert candles[0].close_time == 1704101400000 + 60_000 - 1  # 1m bar
    assert str(candles[0].close) == "100"
    assert str(candles[1].close) == "101"


async def test_fetch_bars_crypto_path_and_headers() -> None:
    seen_headers: list[httpx.Headers] = []

    def handler(request: httpx.Request) -> httpx.Response:
        seen_headers.append(request.headers)
        assert (
            request.url.raw_path.split(b"?")[0] == b"/v1beta3/crypto/us/BTC%2FUSD/bars"
        )
        assert "feed" not in request.url.params  # no iex feed for crypto
        assert request.url.params["timeframe"] == "1Hour"
        return httpx.Response(
            200,
            json={
                "bars": [bar("2024-01-01T09:00:00Z", "50000")],
                "next_page_token": "",
            },
        )

    client = make_client(handler)
    client.headers["APCA-API-KEY-ID"] = "key"
    client.headers["APCA-API-SECRET-KEY"] = "secret"
    candles = await fetch_bars(
        client, "BTC/USD", "1h", 1704067200000, 1704067200000 + 24 * 3_600_000
    )
    assert len(candles) == 1
    assert candles[0].open_time == 1704099600000  # 2024-01-01T09:00:00Z
    assert candles[0].close_time == 1704099600000 + 3_600_000 - 1  # 1h bar
    assert str(candles[0].close) == "50000"
    assert seen_headers[0]["APCA-API-KEY-ID"] == "key"


async def test_fetch_bars_unsupported_interval() -> None:
    with pytest.raises(ValueError, match="unsupported interval"):
        await fetch_bars(
            make_client(lambda r: httpx.Response(200, json={})), "AAPL", "3m", 0, 1
        )


async def test_fetch_bars_http_error_raises() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(403, json={"message": "forbidden"})

    with pytest.raises(httpx.HTTPStatusError):
        await fetch_bars(make_client(handler), "AAPL", "1m", 0, 1000)

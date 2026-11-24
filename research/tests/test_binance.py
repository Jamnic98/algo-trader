import httpx
import pytest

from app.binance import Candle, fetch_klines


def kline_row(open_time: int, close: str) -> list:
    return [
        open_time,
        close,
        close,
        close,
        close,
        "1.0",
        open_time + 59_999,
        "0",
        0,
        "0",
        "0",
        "0",
    ]


def make_client(handler) -> httpx.AsyncClient:
    return httpx.AsyncClient(
        transport=httpx.MockTransport(handler), base_url="https://api.binance.com"
    )


async def test_fetch_klines_paginates_forward() -> None:
    calls = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(dict(request.url.params))
        start = int(request.url.params["startTime"])
        if len(calls) == 1:
            # full batch of 1000 -> must paginate
            rows = [kline_row(start + i * 60_000, "100") for i in range(1000)]
        else:
            rows = [kline_row(start + i * 60_000, "101") for i in range(10)]
        return httpx.Response(200, json=rows)

    client = make_client(handler)
    end = 1_700_000_000_000 + 1010 * 60_000
    candles = await fetch_klines(client, "BTCUSDT", "1m", 1_700_000_000_000, end)

    assert len(candles) == 1010
    assert len(calls) == 2
    assert calls[0]["limit"] == "1000"
    assert calls[0]["endTime"] == str(end)
    # second page starts right after the last openTime of page 1
    assert calls[1]["startTime"] == str(1_700_000_000_000 + 999 * 60_000 + 1)
    assert isinstance(candles[0], Candle)
    assert str(candles[-1].close) == "101"


async def test_fetch_klines_stops_on_empty() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json=[])

    client = make_client(handler)
    candles = await fetch_klines(
        client, "BTCUSDT", "1h", 1_700_000_000_000, 1_700_100_000_000
    )
    assert candles == []


async def test_fetch_klines_rejects_bad_interval() -> None:
    client = make_client(lambda request: httpx.Response(200, json=[]))
    with pytest.raises(ValueError, match="unsupported interval"):
        await fetch_klines(client, "BTCUSDT", "7m", 0, 1000)

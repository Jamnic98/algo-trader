from dataclasses import dataclass
from decimal import Decimal

import httpx

# mirrors engine.ParseInterval on the Go side plus the full Binance set
INTERVALS = {
    "1m",
    "3m",
    "5m",
    "15m",
    "30m",
    "1h",
    "2h",
    "4h",
    "6h",
    "8h",
    "12h",
    "1d",
    "3d",
    "1w",
    "1M",
}

MAX_PER_REQUEST = 1000


def normalize_symbol(symbol: str) -> str:
    """Uppercase and strip separators: btc/usd -> BTCUSDT-style venue symbol."""
    return "".join(c for c in symbol.upper() if c.isalnum())


@dataclass
class Candle:
    open_time: int  # ms
    close_time: int  # ms
    open: Decimal
    high: Decimal
    low: Decimal
    close: Decimal
    volume: Decimal


async def fetch_klines(
    client: httpx.AsyncClient,
    symbol: str,
    interval: str,
    start_ms: int,
    end_ms: int,
) -> list[Candle]:
    """Fetch klines with openTime in [start_ms, end_ms], paginating forward."""
    if interval not in INTERVALS:
        raise ValueError(f"unsupported interval: {interval!r}")

    candles: list[Candle] = []
    cursor = start_ms

    while True:
        resp = await client.get(
            "/api/v3/klines",
            params={
                "symbol": symbol,
                "interval": interval,
                "limit": MAX_PER_REQUEST,
                "startTime": cursor,
                "endTime": end_ms,
            },
        )
        resp.raise_for_status()
        rows = resp.json()
        if not rows:
            break

        for r in rows:
            candles.append(
                Candle(
                    open_time=int(r[0]),
                    close_time=int(r[6]),
                    open=Decimal(str(r[1])),
                    high=Decimal(str(r[2])),
                    low=Decimal(str(r[3])),
                    close=Decimal(str(r[4])),
                    volume=Decimal(str(r[5])),
                )
            )

        last_open = int(rows[-1][0])
        if last_open >= end_ms or len(rows) < MAX_PER_REQUEST:
            break
        cursor = last_open + 1

    return candles

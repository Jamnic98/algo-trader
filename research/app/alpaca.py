import re
from datetime import datetime, timezone
from decimal import Decimal
from urllib.parse import quote

import httpx

from .binance import Candle

# maps our interval strings to Alpaca bar timeframes (subset Alpaca supports)
TIMEFRAMES = {
    "1m": ("1Min", 60_000),
    "5m": ("5Min", 300_000),
    "15m": ("15Min", 900_000),
    "30m": ("30Min", 1_800_000),
    "1h": ("1Hour", 3_600_000),
    "4h": ("4Hour", 14_400_000),
    "1d": ("1Day", 86_400_000),
    "1w": ("1Week", 604_800_000),
}

MAX_PER_REQUEST = 10000


def normalize_symbol(symbol: str) -> str:
    """Uppercase, keep the crypto slash: btc/usd -> BTC/USD, aapl -> AAPL."""
    return symbol.upper()


def is_crypto(symbol: str) -> bool:
    return "/" in symbol


def bars_path(symbol: str) -> str:
    if is_crypto(symbol):
        return f"/v1beta3/crypto/us/{quote(symbol, safe='')}/bars"
    return f"/v2/stocks/{quote(symbol, safe='')}/bars"


def rfc3339(ms: int) -> str:
    return datetime.fromtimestamp(ms / 1000, tz=timezone.utc).strftime(
        "%Y-%m-%dT%H:%M:%SZ"
    )


_BAR_TS = re.compile(
    r"^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.\d+)?(Z|[+-]\d{2}:?\d{2})?$"
)


def parse_bar_time(t: str) -> int:
    """Parse an Alpaca RFC3339 bar timestamp (optional fractional seconds) to unix ms."""
    m = _BAR_TS.match(t)
    if not m:
        raise ValueError(f"bad bar time {t!r}")
    tz = m.group(2) or "Z"
    if tz == "Z":
        tz = "+00:00"
    elif ":" not in tz:
        tz = tz[:3] + ":" + tz[3:]
    return int(datetime.fromisoformat(m.group(1) + tz).timestamp() * 1000)


async def fetch_bars(
    client: httpx.AsyncClient,
    symbol: str,
    interval: str,
    start_ms: int,
    end_ms: int,
) -> list[Candle]:
    """Fetch Alpaca bars with open time in [start_ms, end_ms], paginating via page_token."""
    if interval not in TIMEFRAMES:
        raise ValueError(
            f"unsupported interval for alpaca: {interval!r} "
            f"(supported: {sorted(TIMEFRAMES)})"
        )
    timeframe, interval_ms = TIMEFRAMES[interval]
    symbol = normalize_symbol(symbol)

    candles: list[Candle] = []
    page_token = ""

    while True:
        params: dict[str, str] = {
            "timeframe": timeframe,
            "start": rfc3339(start_ms),
            "end": rfc3339(end_ms),
            "limit": str(MAX_PER_REQUEST),
        }
        if page_token:
            params["page_token"] = page_token
        if not is_crypto(symbol):
            params["feed"] = "iex"  # free tier

        resp = await client.get(bars_path(symbol), params=params)
        resp.raise_for_status()
        page = resp.json()
        bars = page.get("bars") or []
        if not bars:
            break

        for b in bars:
            open_time = parse_bar_time(b["t"])
            candles.append(
                Candle(
                    open_time=open_time,
                    close_time=open_time + interval_ms - 1,
                    open=Decimal(str(b["o"])),
                    high=Decimal(str(b["h"])),
                    low=Decimal(str(b["l"])),
                    close=Decimal(str(b["c"])),
                    volume=Decimal(str(b["v"])),
                )
            )

        page_token = page.get("next_page_token") or ""
        if not page_token:
            break

    return candles

#!/usr/bin/env python3
"""Sentiment strategy service for the "http" strategy template.

Scores recent news headlines from an RSS/Atom feed with a small built-in
lexicon, maps the average score to BUY/SELL/HOLD, and exposes the standard
http-template contract:

    POST /predict  {"closes": [...], "params": {...}}
                -> {"signal": "BUY"|"SELL"|"HOLD", "quantity?": "...", "reason": "..."}

Run locally:
    python3 scripts/sentiment_service.py     # serves http://127.0.0.1:8002

In docker compose the `sentiment` service runs this on :8002 (set HOST=0.0.0.0,
PORT=8002); point the seeded `sentiment` strategy at it via SENTIMENT_URL
(e.g. http://sentiment:8002/predict), which trader-core seeds into the
strategy's default_config at boot.

Params (all optional; the http strategy forwards everything except url/window):
    feed_url        RSS/Atom feed of headlines (default: CoinDesk RSS)
    threshold       |score| above this -> BUY/SELL (default 0.05)
    max_items       max headlines scored (default 25)
    cache_seconds   TTL cache for the fetched feed (default 300; 0 = no cache)
    momentum_weight blend sentiment with last-close momentum, 0..1 (default 0)
    quantity        fixed qty returned on BUY (default: caller decides)
"""

import os
import re
import time
import xml.etree.ElementTree as ET

import httpx
import uvicorn
from fastapi import FastAPI
from pydantic import BaseModel

DEFAULT_FEED_URL = "https://www.coindesk.com/arc/outboundfeeds/rss/"
DEFAULT_THRESHOLD = 0.05
DEFAULT_MAX_ITEMS = 25
DEFAULT_CACHE_SECONDS = 300

# Tiny lexicon: word -> weight in [-1, 1]. Deliberately small and transparent —
# swap in an embedding/LLM scorer behind the same score_titles() signature.
LEXICON = {
    # positive
    "surge": 1.0,
    "surges": 1.0,
    "rally": 1.0,
    "rallies": 1.0,
    "bullish": 1.0,
    "gain": 0.7,
    "gains": 0.7,
    "profit": 0.7,
    "profits": 0.7,
    "record": 0.6,
    "adoption": 0.8,
    "approve": 0.9,
    "approved": 1.0,
    "approval": 1.0,
    "etf": 0.6,
    "upgrade": 0.8,
    "upgraded": 0.8,
    "beat": 0.7,
    "beats": 0.7,
    "growth": 0.7,
    "partnership": 0.5,
    "high": 0.4,
    "soar": 1.0,
    "soars": 1.0,
    "breakout": 0.8,
    "accumulate": 0.6,
    "buy": 0.5,
    "outperform": 0.8,
    # negative
    "crash": -1.0,
    "plunge": -1.0,
    "plunges": -1.0,
    "dump": -0.9,
    "bearish": -1.0,
    "loss": -0.7,
    "losses": -0.7,
    "hack": -1.0,
    "hacked": -1.0,
    "ban": -0.8,
    "banned": -0.9,
    "lawsuit": -0.8,
    "fraud": -1.0,
    "sell-off": -0.9,
    "fear": -0.7,
    "warning": -0.6,
    "warns": -0.6,
    "downgrade": -0.8,
    "downgraded": -0.8,
    "miss": -0.6,
    "misses": -0.6,
    "decline": -0.7,
    "declines": -0.7,
    "fine": -0.6,
    "fined": -0.9,
    "sec": -0.5,
    "regulation": -0.5,
    "low": -0.4,
    "sell": -0.5,
    "drop": -0.7,
    "drops": -0.7,
    "bankruptcy": -1.0,
    "default": -0.8,
}

_WORD_RE = re.compile(r"[a-z']+")

app = FastAPI(title="sentiment-service")


class PredictRequest(BaseModel):
    closes: list[str]
    params: dict = {}


class PredictResponse(BaseModel):
    signal: str
    quantity: str | None = None
    reason: str = ""


def score_text(text: str) -> float:
    """Average lexicon weight of the matched words, in [-1, 1]; 0 if none match."""
    matched = [LEXICON[w] for w in _WORD_RE.findall(text.lower()) if w in LEXICON]
    if not matched:
        return 0.0
    return sum(matched) / len(matched)


def parse_feed_titles(xml_bytes: bytes, max_items: int) -> list[str]:
    """Extract item/entry titles from RSS 2.0 or Atom feeds."""
    root = ET.fromstring(xml_bytes)
    titles: list[str] = []
    # RSS 2.0: <channel><item><title>; Atom: <feed><entry><title> (namespaced)
    for el in root.iter():
        tag = el.tag.rsplit("}", 1)[-1]  # strip namespace
        if tag not in ("item", "entry"):
            continue
        for child in el:
            if child.tag.rsplit("}", 1)[-1] == "title" and child.text:
                titles.append(child.text.strip())
                break
        if len(titles) >= max_items:
            break
    return titles[:max_items]


def _fetch_titles(feed_url: str, max_items: int) -> list[str]:
    resp = httpx.get(feed_url, timeout=10.0, follow_redirects=True)
    resp.raise_for_status()
    return parse_feed_titles(resp.content, max_items)


# feed_url -> (fetched_at, titles); simple TTL cache so per-candle calls don't
# hammer the feed
_feed_cache: dict[str, tuple[float, list[str]]] = {}


def get_titles(feed_url: str, max_items: int, cache_seconds: float) -> list[str]:
    now = time.time()
    if cache_seconds > 0 and feed_url in _feed_cache:
        ts, titles = _feed_cache[feed_url]
        if now - ts < cache_seconds:
            return titles
    titles = _fetch_titles(feed_url, max_items)
    _feed_cache[feed_url] = (now, titles)
    return titles


def sentiment_score(titles: list[str]) -> float:
    """Mean headline score in [-1, 1]; 0 when there are no headlines."""
    if not titles:
        return 0.0
    return sum(score_text(t) for t in titles) / len(titles)


def _num(params: dict, key: str, default: float) -> float:
    v = params.get(key)
    if v is None or v == "":
        return default
    return float(v)


@app.post("/predict", response_model=PredictResponse)
def predict(req: PredictRequest) -> PredictResponse:
    p = req.params
    feed_url = str(p.get("feed_url") or DEFAULT_FEED_URL)
    threshold = _num(p, "threshold", DEFAULT_THRESHOLD)
    max_items = int(_num(p, "max_items", DEFAULT_MAX_ITEMS))
    cache_seconds = _num(p, "cache_seconds", DEFAULT_CACHE_SECONDS)
    momentum_weight = _num(p, "momentum_weight", 0.0)

    try:
        titles = get_titles(feed_url, max_items, cache_seconds)
    except Exception as exc:  # feed unreachable / unparsable -> safe HOLD
        return PredictResponse(signal="HOLD", reason=f"feed error: {exc}")

    score = sentiment_score(titles)

    # optional blend with last-close momentum (scaled to [-1, 1])
    if momentum_weight > 0 and len(req.closes) >= 2:
        try:
            prev, last = float(req.closes[-2]), float(req.closes[-1])
            if prev > 0:
                mom = max(-1.0, min(1.0, (last / prev - 1.0) * 100.0))
                score = (1.0 - momentum_weight) * score + momentum_weight * mom
        except (ValueError, ZeroDivisionError):
            pass

    n = len(titles)
    if score > threshold:
        qty = p.get("quantity")
        return PredictResponse(
            signal="BUY",
            quantity=str(qty) if qty not in (None, "") else None,
            reason=f"score={score:.4f} over {n} headlines",
        )
    if score < -threshold:
        return PredictResponse(
            signal="SELL", reason=f"score={score:.4f} over {n} headlines"
        )
    return PredictResponse(
        signal="HOLD", reason=f"score={score:.4f} over {n} headlines"
    )


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


if __name__ == "__main__":
    uvicorn.run(
        app,
        host=os.environ.get("HOST", "127.0.0.1"),
        port=int(os.environ.get("PORT", "8002")),
    )

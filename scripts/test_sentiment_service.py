"""Tests for the sentiment http-strategy service (scripts/sentiment_service.py).

Run from the repo root:  python -m pytest scripts/test_sentiment_service.py -q
(CI runs this in the Test Backtester job after installing research/requirements.txt.)
"""

import httpx
import pytest
from fastapi.testclient import TestClient

import sentiment_service as svc

RSS_SAMPLE = """<?xml version="1.0"?><rss version="2.0"><channel><title>feed</title>
<item><title>Bitcoin surges to record high on ETF approval</title></item>
<item><title>Market crash as exchange hack drains funds</title></item>
<item><title>Local council debates park rules</title></item>
</channel></rss>"""

ATOM_SAMPLE = """<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom"><title>feed</title>
<entry><title>Bearish warning from analysts</title></entry>
<entry><title>Prices rally after upgrade</title></entry>
</feed>"""


@pytest.fixture(autouse=True)
def clear_cache():
    svc._feed_cache.clear()
    yield
    svc._feed_cache.clear()


@pytest.fixture
def client() -> TestClient:
    return TestClient(svc.app)


def _patch_titles(monkeypatch: pytest.MonkeyPatch, titles: list[str]) -> list[int]:
    calls = [0]

    def fake_fetch(feed_url: str, max_items: int) -> list[str]:
        calls[0] += 1
        return titles[:max_items]

    monkeypatch.setattr(svc, "_fetch_titles", fake_fetch)
    return calls


# --- scoring ---------------------------------------------------------------


def test_score_text_positive_negative_neutral() -> None:
    assert svc.score_text("Bitcoin surges to a record high") > 0.5
    assert svc.score_text("Market crash, exchange hacked, lawsuit") < -0.5
    assert svc.score_text("Local council debates park rules") == 0.0
    assert svc.score_text("") == 0.0
    # mixed headline lands between
    mixed = svc.score_text("Bitcoin surges despite crash fears")
    assert -1.0 < mixed < 1.0


def test_parse_feed_titles_rss_and_atom() -> None:
    rss_titles = svc.parse_feed_titles(RSS_SAMPLE.encode(), 25)
    assert len(rss_titles) == 3
    assert "Bitcoin surges" in rss_titles[0]
    atom_titles = svc.parse_feed_titles(ATOM_SAMPLE.encode(), 25)
    assert len(atom_titles) == 2
    # max_items caps the list
    assert len(svc.parse_feed_titles(RSS_SAMPLE.encode(), 2)) == 2


def test_sentiment_score_mean() -> None:
    assert svc.sentiment_score([]) == 0.0
    titles = ["Bitcoin surges to record high", "Market crash and hack"]
    s = svc.sentiment_score(titles)
    assert s == pytest.approx(
        (svc.score_text(titles[0]) + svc.score_text(titles[1])) / 2
    )


# --- /predict ---------------------------------------------------------------


def test_predict_buy_on_positive_headlines(client: TestClient, monkeypatch) -> None:
    _patch_titles(
        monkeypatch,
        ["Bitcoin surges to record high on ETF approval", "Rally continues"],
    )
    resp = client.post(
        "/predict", json={"closes": ["100", "101"], "params": {"threshold": 0.05}}
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["signal"] == "BUY"
    assert body["quantity"] is None
    assert "score=" in body["reason"]


def test_predict_sell_on_negative_headlines(client: TestClient, monkeypatch) -> None:
    _patch_titles(
        monkeypatch, ["Market crash as exchange hack drains funds", "Bearish warning"]
    )
    resp = client.post("/predict", json={"closes": ["100", "99"], "params": {}})
    assert resp.json()["signal"] == "SELL"


def test_predict_hold_when_neutral(client: TestClient, monkeypatch) -> None:
    _patch_titles(monkeypatch, ["Local council debates park rules"])
    resp = client.post("/predict", json={"closes": ["100", "100"], "params": {}})
    assert resp.json()["signal"] == "HOLD"


def test_predict_hold_on_feed_error(client: TestClient, monkeypatch) -> None:
    def boom(feed_url: str, max_items: int) -> list[str]:
        raise httpx.ConnectError("no network")

    monkeypatch.setattr(svc, "_fetch_titles", boom)
    resp = client.post("/predict", json={"closes": ["100", "101"], "params": {}})
    assert resp.status_code == 200
    body = resp.json()
    assert body["signal"] == "HOLD"
    assert "feed error" in body["reason"]


def test_predict_quantity_passthrough(client: TestClient, monkeypatch) -> None:
    _patch_titles(monkeypatch, ["Bitcoin surges to record high"])
    resp = client.post(
        "/predict", json={"closes": ["100"], "params": {"quantity": "0.25"}}
    )
    assert resp.json()["signal"] == "BUY"
    assert resp.json()["quantity"] == "0.25"


def test_predict_uses_feed_cache(client: TestClient, monkeypatch) -> None:
    calls = _patch_titles(monkeypatch, ["Bitcoin surges to record high"])
    params = {"cache_seconds": 300}
    client.post("/predict", json={"closes": ["100"], "params": params})
    client.post("/predict", json={"closes": ["101"], "params": params})
    assert calls[0] == 1  # second call served from cache
    client.post("/predict", json={"closes": ["102"], "params": {"cache_seconds": 0}})
    assert calls[0] == 2  # cache disabled -> refetch


def test_predict_momentum_blend(client: TestClient, monkeypatch) -> None:
    _patch_titles(monkeypatch, ["Local council debates park rules"])  # sentiment 0
    # pure momentum: +10% last candle -> mom = 1 -> BUY
    resp = client.post(
        "/predict",
        json={
            "closes": ["100", "110"],
            "params": {"momentum_weight": 1, "threshold": 0.05},
        },
    )
    assert resp.json()["signal"] == "BUY"
    # and -10% -> SELL
    resp = client.post(
        "/predict",
        json={
            "closes": ["100", "90"],
            "params": {"momentum_weight": 1, "threshold": 0.05},
        },
    )
    assert resp.json()["signal"] == "SELL"


def test_health(client: TestClient) -> None:
    assert client.get("/health").json() == {"status": "ok"}

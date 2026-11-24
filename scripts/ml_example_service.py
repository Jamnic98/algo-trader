#!/usr/bin/env python3
"""Example ML strategy service for the "http" template.

A real model would load weights here; this toy "model" is a linear rule on the
last return, exposed over HTTP exactly like an inference endpoint would be.

Run:
    python3 scripts/ml_example_service.py     # serves http://127.0.0.1:8001

Then backtest the seeded `ml_example` strategy (url defaults to
http://127.0.0.1:8001/predict) — from the dashboard or:

    curl -X POST localhost:8000/backtest -H 'Content-Type: application/json' -d '{
      "strategy_slug": "ml_example", "symbol": "BTCUSDT", "interval": "1h",
      "start": "2024-01-01", "end": "2024-03-01" }'
"""

import uvicorn
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="ml-example")

# toy linear model: signal from last log-ish return, weighted
W_RET = 500.0
THRESH = 0.25


class PredictRequest(BaseModel):
    closes: list[str]
    params: dict = {}


class PredictResponse(BaseModel):
    signal: str
    quantity: str | None = None
    reason: str = ""


@app.post("/predict", response_model=PredictResponse)
def predict(req: PredictRequest) -> PredictResponse:
    closes = [float(c) for c in req.closes]
    if len(closes) < 2:
        return PredictResponse(signal="HOLD", reason="need >= 2 closes")
    ret = closes[-1] / closes[-2] - 1.0
    score = W_RET * ret
    if score > THRESH:
        return PredictResponse(signal="BUY", quantity="0.1", reason=f"score={score:.3f}")
    if score < -THRESH:
        return PredictResponse(signal="SELL", reason=f"score={score:.3f}")
    return PredictResponse(signal="HOLD", reason=f"score={score:.3f}")


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8001)

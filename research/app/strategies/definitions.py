import json
import os
from functools import lru_cache
from pathlib import Path
from typing import Any

# shared/strategies.json is the single source of truth for strategy
# definitions; both this service and trader-core (Go) load it.
_REPO_ROOT = Path(__file__).resolve().parents[3]


def _candidates() -> list[str]:
    env = os.environ.get("STRATEGIES_SCHEMA_PATH")
    if env:
        return [env]
    return [
        "shared/strategies.json",
        "../shared/strategies.json",
        str(_REPO_ROOT / "shared" / "strategies.json"),
    ]


@lru_cache(maxsize=1)
def load_schema() -> dict[str, Any]:
    for path in _candidates():
        try:
            with open(path, encoding="utf-8") as f:
                return json.load(f)
        except OSError:
            continue
    raise FileNotFoundError(
        "shared/strategies.json not found; set STRATEGIES_SCHEMA_PATH "
        f"(tried: {_candidates()})"
    )


def find_primitive(name: str) -> dict[str, Any] | None:
    for defn in load_schema()["primitives"]:
        if defn["name"] == name:
            return defn
    return None


def param_defaults(defn: dict[str, Any]) -> dict[str, Any]:
    return {
        p["key"]: p["default"]
        for p in defn.get("params", [])
        if p.get("default") is not None
    }


def composite_default_threshold() -> float:
    return float(load_schema().get("composite", {}).get("default_threshold") or 0.6)

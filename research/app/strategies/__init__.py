from typing import Any, Callable

from ..store import StrategyRow, StrategyStore
from .base import CandleContext, Decision, Signal, Strategy
from .composer import WeightedComposer
from .dca import SimpleDCAStrategy, SmartDCAStrategy
from .simple import SimpleStrategy

__all__ = [
    "CandleContext",
    "Decision",
    "Signal",
    "Strategy",
    "build_primitive",
    "resolve",
]

PRIMITIVES: dict[str, Callable[[dict[str, Any]], Strategy]] = {
    "simple": SimpleStrategy,
    "simple_dca": SimpleDCAStrategy,
    "smart_dca": SmartDCAStrategy,
}

DEFAULT_THRESHOLD = 0.6


def build_primitive(slug: str, params: dict[str, Any]) -> Strategy:
    cls = PRIMITIVES.get(slug)
    if cls is None:
        raise ValueError(f"unknown strategy {slug!r}")
    return cls(params)


def _params_of(config: dict[str, Any]) -> dict[str, Any]:
    # Go's Build() passes DefaultConfig straight through as the params blob;
    # some configs nest params under "params" (StrategyConfigDTO shape).
    nested = config.get("params")
    return nested if isinstance(nested, dict) else config


def resolve(
    row: StrategyRow, overrides: dict[str, Any] | None, store: StrategyStore
) -> Strategy:
    """Build a Strategy from a DB row, mirroring Go bot.ResolveStrategy."""
    if not row.is_composite:
        config = dict(row.default_config or {})
        if overrides:
            config.update(overrides)
        return build_primitive(row.slug, _params_of(config))

    config = row.default_config or {}
    nodes_cfg = config.get("nodes") or row.nodes or []
    if not nodes_cfg:
        raise ValueError(f"composite strategy {row.slug!r} has no nodes")

    threshold = float(config.get("threshold") or DEFAULT_THRESHOLD)
    nodes: list[tuple[Strategy, float]] = []
    for node in nodes_cfg:
        sub = store.get(node["strategy_slug"])
        if sub is None:
            raise ValueError(f"unknown sub-strategy {node['strategy_slug']!r}")
        nodes.append(
            (resolve(sub, node.get("params"), store), float(node.get("weight", 1)))
        )

    return WeightedComposer(threshold, nodes)

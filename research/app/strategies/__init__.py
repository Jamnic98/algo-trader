from typing import Any

from ..store import StrategyRow, StrategyStore
from .base import CandleContext, Decision, Signal, Strategy
from .composer import WeightedComposer
from .definitions import composite_default_threshold, find_primitive, param_defaults
from .dca import SimpleDCAStrategy, SmartDCAStrategy
from .simple import SimpleStrategy
from .sma_cross import SMACrossStrategy

__all__ = [
    "CandleContext",
    "Decision",
    "Signal",
    "Strategy",
    "build_primitive",
    "resolve",
]

# Templates are declared in shared/strategies.json and implemented in both
# this package and trader-core/internal/strategies (see definitions.go).
# The shared test vectors in strategies.json pin the behavior on both sides.
TEMPLATES: dict[str, type] = {
    "always_buy": SimpleDCAStrategy,
    "prev_close": SimpleStrategy,
    "dca": SmartDCAStrategy,
    "sma_cross": SMACrossStrategy,
}


def build_primitive(slug: str, params: dict[str, Any] | None) -> Strategy:
    defn = find_primitive(slug)
    if defn is None:
        raise ValueError(f"unknown strategy {slug!r}")
    merged = {**param_defaults(defn), **(params or {})}
    cls = TEMPLATES.get(defn["template"])
    if cls is None:
        raise ValueError(f"unknown template {defn['template']!r} for strategy {slug!r}")
    return cls(merged)


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

    threshold = float(config.get("threshold") or composite_default_threshold())
    nodes: list[tuple[Strategy, float]] = []
    for node in nodes_cfg:
        sub = store.get(node["strategy_slug"])
        if sub is None:
            raise ValueError(f"unknown sub-strategy {node['strategy_slug']!r}")
        nodes.append(
            (resolve(sub, node.get("params"), store), float(node.get("weight", 1)))
        )

    return WeightedComposer(threshold, nodes)

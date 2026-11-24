import json
from dataclasses import dataclass, field
from typing import Any, Protocol


@dataclass
class StrategyRow:
    slug: str
    display_name: str
    is_composite: bool
    default_config: dict[str, Any] = field(default_factory=dict)
    nodes: list[dict[str, Any]] = field(default_factory=list)


class StrategyStore(Protocol):
    def get(self, slug: str) -> StrategyRow | None: ...


class PostgresStrategyStore:
    """Reads saved strategies from the shared Postgres `strategies` table."""

    def __init__(self, dsn: str) -> None:
        self.dsn = dsn

    def get(self, slug: str) -> StrategyRow | None:
        import psycopg

        with psycopg.connect(self.dsn) as conn:
            row = conn.execute(
                "SELECT slug, display_name, is_composite, default_config, nodes "
                "FROM strategies WHERE slug = %s",
                (slug,),
            ).fetchone()

        if row is None:
            return None

        slug_, display_name, is_composite, default_config, nodes = row
        return StrategyRow(
            slug=slug_,
            display_name=display_name,
            is_composite=bool(is_composite),
            default_config=_as_dict(default_config),
            nodes=_as_list(nodes),
        )


class StaticStrategyStore:
    """In-memory store for tests and local runs."""

    def __init__(self, rows: dict[str, StrategyRow]) -> None:
        self.rows = rows

    def get(self, slug: str) -> StrategyRow | None:
        return self.rows.get(slug)


def _as_dict(value: Any) -> dict[str, Any]:
    if value is None:
        return {}
    if isinstance(value, (dict, list)):
        return value if isinstance(value, dict) else json.loads(json.dumps(value))
    return json.loads(value)


def _as_list(value: Any) -> list[dict[str, Any]]:
    if value is None:
        return []
    if isinstance(value, list):
        return value
    return json.loads(value)

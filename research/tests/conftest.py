import os

import pytest

os.environ.setdefault("APP_ENV", "test")

from app.store import StaticStrategyStore, StrategyRow  # noqa: E402


@pytest.fixture
def store() -> StaticStrategyStore:
    return StaticStrategyStore(
        {
            "simple": StrategyRow(
                slug="simple",
                display_name="Simple",
                is_composite=False,
                default_config={},
            ),
            "simple_dca": StrategyRow(
                slug="simple_dca",
                display_name="Simple DCA",
                is_composite=False,
                default_config={},
            ),
            "smart_dca": StrategyRow(
                slug="smart_dca",
                display_name="Smart DCA",
                is_composite=False,
                default_config={"fixed_spend": "10.00"},
            ),
            "combo": StrategyRow(
                slug="combo",
                display_name="Combo",
                is_composite=True,
                default_config={
                    "threshold": 0.6,
                    "nodes": [
                        {"strategy_slug": "simple_dca", "weight": 1.0},
                        {"strategy_slug": "simple", "weight": 1.0},
                    ],
                },
                nodes=[],
            ),
        }
    )

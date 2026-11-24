from app.store import StaticStrategyStore, StrategyRow, _as_dict, _as_list


def test_static_store_hit_and_miss() -> None:
    store = StaticStrategyStore(
        {
            "simple": StrategyRow(
                slug="simple", display_name="Simple", is_composite=False
            )
        }
    )
    row = store.get("simple")
    assert row is not None and row.slug == "simple"
    assert store.get("nope") is None


def test_as_dict_handles_none_str_dict() -> None:
    assert _as_dict(None) == {}
    assert _as_dict('{"a": 1}') == {"a": 1}
    assert _as_dict({"a": 1}) == {"a": 1}


def test_as_list_handles_none_str_list() -> None:
    assert _as_list(None) == []
    assert _as_list('[{"strategy_slug": "simple"}]') == [{"strategy_slug": "simple"}]
    assert _as_list([{"strategy_slug": "simple"}]) == [{"strategy_slug": "simple"}]

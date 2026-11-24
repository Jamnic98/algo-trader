import pytest

from app.timeutil import parse_time_ms


def test_unix_ms_int_and_str() -> None:
    assert parse_time_ms(1_700_000_000_000) == 1_700_000_000_000
    assert parse_time_ms("1700000000000") == 1_700_000_000_000


def test_rfc3339() -> None:
    assert parse_time_ms("2024-01-01T00:00:00Z") == 1704067200000
    assert parse_time_ms("2024-01-01T02:00:00+02:00") == 1704067200000


def test_date_only_is_utc_midnight() -> None:
    assert parse_time_ms("2024-01-01") == 1704067200000


def test_invalid() -> None:
    with pytest.raises(ValueError):
        parse_time_ms("not-a-date")
    with pytest.raises(ValueError):
        parse_time_ms("2024-13-45")

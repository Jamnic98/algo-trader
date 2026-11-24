from datetime import datetime, timezone


def parse_time_ms(value: str | int) -> int:
    """Accept unix milliseconds, RFC3339 or YYYY-MM-DD (UTC). Returns ms."""
    if isinstance(value, int):
        return value
    s = value.strip()
    if s.isdigit():
        return int(s)
    try:
        return int(datetime.fromisoformat(s.replace("Z", "+00:00")).timestamp() * 1000)
    except ValueError:
        pass
    try:
        dt = datetime.strptime(s, "%Y-%m-%d").replace(tzinfo=timezone.utc)
        return int(dt.timestamp() * 1000)
    except ValueError as exc:
        raise ValueError(
            f"expected unix ms, RFC3339 or YYYY-MM-DD, got {value!r}"
        ) from exc

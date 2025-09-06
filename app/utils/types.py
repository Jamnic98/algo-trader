from typing import Any, TypedDict

from sqlalchemy.pool import Pool


class EngineOptions(TypedDict, total=False):
    future: bool
    pool_pre_ping: bool
    echo: bool
    connect_args: dict[str, Any]
    poolclass: type[Pool]

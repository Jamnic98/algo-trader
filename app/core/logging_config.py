import logging
import sys
from typing import Optional

from app.core.config import settings

_INITIALISED: bool = False


def init_logging(logfile: Optional[str] = None) -> None:
    global _INITIALISED
    if _INITIALISED:
        return

    for h in list(logging.root.handlers):
        logging.root.removeHandler(h)

    handlers: list[logging.Handler] = [logging.StreamHandler(sys.stdout)]
    if logfile:
        handlers.append(logging.FileHandler(logfile, mode="a", encoding="utf-8"))

    logging.basicConfig(
        level=logging.DEBUG if settings.debug else logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
        datefmt="%H:%M:%S",
        handlers=handlers,
        force=True,
    )
    _INITIALISED = True


def get_logger(name: str) -> logging.Logger:
    return logging.getLogger(name)


def setup_logging(logfile: Optional[str] = None) -> logging.Logger:
    init_logging(logfile=logfile)
    return logging.getLogger("algo-trader")

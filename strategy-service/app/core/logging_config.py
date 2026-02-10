import logging
import sys
from app.core.config import settings

_INITIALISED: bool = False


def init_logging() -> None:
    global _INITIALISED
    if _INITIALISED:
        return

    # remove any existing handlers
    for h in list(logging.root.handlers):
        logging.root.removeHandler(h)

    # only log to stdout
    handlers = [logging.StreamHandler(sys.stdout)]

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


def setup_logging() -> logging.Logger:
    init_logging()
    return logging.getLogger("algo-trader")

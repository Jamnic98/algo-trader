import os
import sys

from src.utils.logging import logger


def get_env_var(key: str, required: bool = True) -> str:
    """Fetch an environment variable and ensure it's not None if required."""
    value = os.getenv(key)
    if required and not value:
        logger.error("Missing required environment variable %s", key)
        sys.exit(1)
    return value

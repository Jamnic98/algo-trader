from app.core.config import settings
from app.core.logging_config import setup_logging

logger = setup_logging(__name__)


def main() -> None:
    logger.info("Starting application")
    logger.info(f"Environment: {settings.app_env}")


if __name__ == "__main__":
    main()

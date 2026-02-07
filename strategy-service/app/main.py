from uvicorn import run

from app.core.config import settings
from app.core.logging_config import setup_logging
from app.factories.fastapi_app import create_app

logger = setup_logging(__name__)

app = create_app()


def main() -> None:
    logger.info(f"Environment: {settings.app_env}")
    run(app)


if __name__ == "__main__":
    main()

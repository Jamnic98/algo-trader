from app.core.logging_config import setup_logging

logger = setup_logging()


def main() -> None:
    logger.info(f"Starting {__name__}")


if __name__ == "__main__":
    main()

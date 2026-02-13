import os


def pytest_configure() -> None:
    os.environ.setdefault("APP_ENV", "test")

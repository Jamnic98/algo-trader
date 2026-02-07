import os

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    debug: bool = True
    db_url: str = "sqlite:///./dev.db"
    app_env: str = os.getenv("APP_ENV", "local").lower()

    model_config = SettingsConfigDict(
        env_file=f".env.{app_env}",
        env_file_encoding="utf-8",
    )


settings = Settings()

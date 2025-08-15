import os

from pydantic_settings import BaseSettings, SettingsConfigDict

env_type = os.getenv("APP_ENV", "local").lower()
env_file = ".env" if env_type == "local" else f".env.{env_type}"


class Settings(BaseSettings):
    debug: bool = True

    model_config = SettingsConfigDict(env_file=env_file, env_file_encoding="utf-8")


settings = Settings()

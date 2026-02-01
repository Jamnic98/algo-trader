import os

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    debug: bool = True
    db_url: str = "sqlite:///./dev.db"
    app_env: str = os.getenv("APP_ENV", "local").lower()

    model_config = SettingsConfigDict(
        env_file=(
            f".env.{os.getenv('APP_ENV', 'local').lower()}"
            if os.getenv("APP_ENV", "local").lower() != "local"
            else ".env"
        ),
        env_file_encoding="utf-8",
    )

    @property
    def effective_db_url(self) -> str:
        """Return the DB URL depending on environment."""
        if self.app_env == "test":
            return "sqlite:///:memory:"

        return self.db_url


settings = Settings()

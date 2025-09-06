from sqlalchemy import create_engine
from sqlalchemy.orm import Session, declarative_base, sessionmaker
from sqlalchemy.pool import StaticPool

from app.core.config import settings
from app.utils.types import EngineOptions

# Shared declarative base for models
Base = declarative_base()

# Default engine options
engine_options: EngineOptions = {"future": True, "pool_pre_ping": True}

if settings.app_env == "local":
    engine_options.update({"echo": True})
elif settings.app_env == "test":
    # Use StaticPool so in-memory SQLite persists across connections in tests
    engine_options.update(
        {
            "connect_args": {"check_same_thread": False},
            "poolclass": StaticPool,
        }
    )

engine = create_engine(settings.effective_db_url, **engine_options)

SessionLocal: sessionmaker[Session] = sessionmaker(
    autocommit=False, autoflush=False, bind=engine, future=True
)

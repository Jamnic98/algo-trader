from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import declarative_base
from sqlalchemy.pool import NullPool

from app.core.config import settings

Base = declarative_base()

db_engine = create_async_engine(settings.db_url, poolclass=NullPool)

async_sessionmaker_instance = async_sessionmaker(
    bind=db_engine,
    expire_on_commit=False,
    class_=AsyncSession,
)


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    async with async_sessionmaker_instance() as session:
        yield session

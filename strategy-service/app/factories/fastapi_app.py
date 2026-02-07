from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, async_sessionmaker

from .db import Base, async_sessionmaker_instance, db_engine


async def async_create_tables(engine: AsyncEngine) -> None:
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


def get_lifespan(engine: AsyncEngine):  # type: ignore
    @asynccontextmanager
    async def lifespan(_app: FastAPI) -> AsyncGenerator[None, None]:
        await async_create_tables(engine)
        yield

    return lifespan


def create_app(
    engine: AsyncEngine | None = None,
    asyncsessionmaker: async_sessionmaker[AsyncSession] | None = None,
) -> FastAPI:
    if engine is None or asyncsessionmaker is None:
        engine, asyncsessionmaker = db_engine, async_sessionmaker_instance

    app = FastAPI(lifespan=get_lifespan(engine))

    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    @app.get("/health")
    def health() -> dict[str, str]:
        return {"status": "ok"}

    return app

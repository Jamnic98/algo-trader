import logging

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routes.decisions import router as decisions_router

logger = logging.getLogger(__name__)


def create_app() -> FastAPI:
    app = FastAPI()

    app.add_middleware(
        CORSMiddleware,
        allow_origins=["http://localhost"],
        allow_methods=["*"],
        allow_headers=["*"],
        allow_credentials=True,
    )

    app.include_router(decisions_router, prefix="/decisions")

    @app.get("/health/")
    def health() -> dict[str, str]:
        return {"status": "ok"}

    return app

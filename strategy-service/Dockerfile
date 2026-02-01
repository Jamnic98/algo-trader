FROM python:3.13-slim AS base

WORKDIR /app
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

RUN pip install --no-cache-dir uv

COPY pyproject.toml uv.lock ./

ARG INSTALL_DEV=false
RUN if [ "$INSTALL_DEV" = "true" ]; then \
        uv sync --frozen --dev; \
    else \
        uv sync --frozen --no-dev; \
    fi

COPY . .

EXPOSE 8000
CMD ["uv", "run", "python", "main.py"]

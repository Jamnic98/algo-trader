# Algo Trader

A minimal algorithmic trading bot prototype, designed for rapid development and environment flexibility.

## Features

- Environment-based settings using `pydantic-settings`
- Simple logging configuration
- Docker-ready with multi-environment support
- Dev tooling: `black`, `ruff`, `mypy`, `pytest`
- Built with Python 3.13

---

## **Local Development Setup**

### Install dependencies
This project uses [uv](https://github.com/astral-sh/uv) as the package/dependency manager.

```bash
# Install uv (if not installed)
pip install uv

# Install project dependencies (including dev tools)
uv sync --dev
```

---

## **Running Locally**

### Run the app
```bash
  uv run python main.py
```

---

## **Docker Usage**

Build the image once and specify the environment at runtime.

### Build the Docker Image
```bash
  docker build --build-arg INSTALL_DEV=true -t algo-trader .
```

### Run the Container

**Local**
```bash
  docker run algo-trader
```

**Development**
```bash
  docker run -e APP_ENV=dev algo-trader
```

**Production**
```bash
  docker run -e APP_ENV=prod algo-trader
```

---

## **Environment Configuration**

The environment variable `APP_ENV` determines which `.env` file is loaded.

| APP_ENV value | Loaded env file |
|---------------|-----------------|
| `local`       | `.env`          |
| `dev`         | `.env.dev`      |
| `prod`        | `.env.prod`     |

Example `.env` file for local development:
```env
DEBUG=true
```

---

## **Project Structure**

```
algo-trader/
├── app/
│   ├── core/
│   │   ├── logging_config.py
│   │   └── settings.py
│   └── ...
├── scripts/
│   └── dev-check.sh
├── tests/
├── main.py
├── pyproject.toml
└── README.md
```

---

## **Development Utilities**

The `dev-check.sh` script runs code formatting, linting, and type checks.  
It’s optional and only needed during development.

### Make the script executable (run once)
```bash
  chmod +x ./scripts/dev-check.sh
```

### Run the script
```bash
  ./scripts/dev-check.sh
```

---

## **License**

MIT License.

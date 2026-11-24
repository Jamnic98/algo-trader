import os

DB_DSN = os.environ.get("DB_DSN", "")
BINANCE_REST_BASE = os.environ.get("BINANCE_REST_BASE", "https://api.binance.com")

# Alpaca market data (bars). Paper vs live only affects the trading API; the
# market-data host is the same. Credentials are required for the data API.
ALPACA_DATA_URL = os.environ.get("ALPACA_DATA_URL", "https://data.alpaca.markets")
ALPACA_API_KEY = os.environ.get("ALPACA_API_KEY", "")
ALPACA_SECRET = os.environ.get("ALPACA_SECRET", "")

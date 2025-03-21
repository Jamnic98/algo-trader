from datetime import datetime
from alpaca.data.historical.crypto import CryptoHistoricalDataClient
from alpaca.data.requests import CryptoBarsRequest
from alpaca.data.timeframe import TimeFrame
from src.utils.settings import config

crypto_data_client = CryptoHistoricalDataClient(config['ALPACA_KEY'], config['ALPACA_SECRET'])

# Specify the request to fetch crypto bars
request = CryptoBarsRequest(
    symbol_or_symbols='BTC/USD',
    timeframe=TimeFrame.Hour,
    start=datetime(2025, 1, 1),  # Example start date
    end=datetime(2025, 1, 31)  # Example end date
)

# Get crypto bars
bars = crypto_data_client.get_crypto_bars(request)

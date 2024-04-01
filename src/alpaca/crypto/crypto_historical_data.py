""" fetches crypto historical data """

import pandas as pd
import plotly.graph_objects as go
from alpaca.data.historical.crypto import CryptoHistoricalDataClient
from alpaca.data.requests import CryptoBarsRequest
from alpaca.data.timeframe import TimeFrame
from src.settings import config

crypto_data_client = CryptoHistoricalDataClient(config['ALPACA_KEY'], config['ALPACA_SECRET'])

# Specify the request to fetch crypto bars
request = CryptoBarsRequest(
    symbol_or_symbols='BTC/USD',
    timeframe=TimeFrame.Hour,
    start="2024-03-01",  # Example start date
    # end="2023-01-31"     # Example end date
)

# Get crypto bars
bars = crypto_data_client.get_crypto_bars(request)
df = pd.DataFrame(bars.df)
df = df.reset_index([0, 1]).drop(['symbol'], axis=1)

fig = go.Figure(data=[go.Candlestick(x=df['timestamp'],
                open=df['open'],
                high=df['high'],
                low=df['low'],
                close=df['close'])])

fig.show()

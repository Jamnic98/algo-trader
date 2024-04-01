""" fetches alpaca historical stock data """

import pandas as pd
import plotly.graph_objects as go
from alpaca.data.historical.stock import StockHistoricalDataClient
from alpaca.data.requests import StockBarsRequest
from alpaca.data.timeframe import TimeFrame
from src.settings import config

stock_data_client = StockHistoricalDataClient(config['ALPACA_KEY'], config['ALPACA_SECRET'])

# Specify the request to fetch stock bars
request = StockBarsRequest(
    symbol_or_symbols='AAPL',
    timeframe=TimeFrame.Hour,
    start="2024-03-01",  # Example start date
    # end="2023-01-31"     # Example end date
)

# Get stock bars
bars = stock_data_client.get_stock_bars(request)
df = pd.DataFrame(bars.df)
df = df.reset_index([0, 1]).drop(['symbol'], axis=1)

fig = go.Figure(data=[go.Candlestick(x=df['timestamp'],
                open=df['open'],
                high=df['high'],
                low=df['low'],
                close=df['close'])])

fig.show()

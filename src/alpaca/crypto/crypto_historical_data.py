import numpy as np
import pandas as pd
import plotly.graph_objects as go
import plotly.express as px
from alpaca.data.historical.crypto import CryptoHistoricalDataClient
from alpaca.data.requests import CryptoBarsRequest, CryptoTradesRequest
from alpaca.data.timeframe import TimeFrame
from src.settings import config

crypto_data_client = CryptoHistoricalDataClient(config['ALPACA_KEY'], config['ALPACA_SECRET'])

# Specify the request to fetch crypto bars
request = CryptoBarsRequest(
    symbol_or_symbols='BTC/USD',
    timeframe=TimeFrame.Hour,
    start="2024-01-01",  # Example start date
    # end="2023-01-31"     # Example end date
)

# Get crypto bars
bars = crypto_data_client.get_crypto_bars(request)
df = pd.DataFrame(bars.df)
df = df.reset_index([0, 1]).drop('symbol', axis=1)

window_1 = 50
window_2 = 252

df['sma1'] = df['close'].rolling(window=window_1).mean()
df['sma2'] = df['close'].rolling(window=window_2).mean()
df['positions'] = np.where(df['sma1'] > df['sma2'], 1, -1)


fig = go.Figure()

# Add traces
# fig.add_hline(df['timestamp'], y=df['positions'])


fig.add_trace(go.Scatter(x=df['timestamp'], y=df['sma1'], mode='lines',))
fig.add_trace(go.Scatter(x=df['timestamp'], y=df['sma2'], mode='lines',))

# Add trace for positions on secondary y-axis
fig.add_trace(
    go.Scatter(
        x=df['timestamp'],
        y=df['positions'],
        marker={'color': 'blue'},
        mode='lines',
        name='Positions',
        yaxis='y2'
    )
)

# Update layout to include a secondary y-axis
fig.update_layout(yaxis2=dict(title='Positions', overlaying='y', side='right'))

fig.add_trace(
    go.Candlestick(
        x=df['timestamp'],
        open=df['open'],
        high=df['high'],
        low=df['low'],
        close=df['close']
    )
)

fig.show()

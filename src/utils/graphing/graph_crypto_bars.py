import numpy as np
import pandas as pd
import plotly.graph_objects as go

def graph_crypto_bars(bars):
    df = pd.DataFrame(bars.df)
    df = df.reset_index([0, 1]).drop('symbol', axis=1)

    window_1 = 50
    window_2 = 252

    df['sma1'] = df['close'].rolling(window=window_1).mean()
    df['sma2'] = df['close'].rolling(window=window_2).mean()
    df['positions'] = np.where(df['sma1'] > df['sma2'], 1, -1)

    fig = go.Figure()

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
    fig.update_layout(yaxis2={'title': 'Positions', 'overlaying': 'y', 'side': 'right'})
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

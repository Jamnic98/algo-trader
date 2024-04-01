from src.settings import config
from alpaca.trading.client import TradingClient

trading_client = TradingClient(
    config['ALPACA_KEY'],
    config['ALPACA_SECRET'],
    paper=not config['LIVE_TRADING']
)

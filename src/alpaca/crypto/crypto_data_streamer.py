from typing import Callable
from alpaca.data.live import CryptoDataStream

from src.utils.settings import config


class CryptoDataStreamer:
    def __init__(self):
        self.wss_client = CryptoDataStream(config['ALPACA_KEY'], config['ALPACA_SECRET'])

    def run(self):
        self.wss_client.run()

    def subscribe_bars(self, subscribe_bars_handler: Callable, ticker_symbols: list = None):
        self.wss_client.subscribe_bars(subscribe_bars_handler, *ticker_symbols)

    def unsubscribe_bars(self, ticker_symbols: list = None):
        self.wss_client.unsubscribe_bars(*ticker_symbols)

# async handler
async def quote_data_handler(data):
    """ handler for stream data """
    # quote data will arrive here
    print(data)

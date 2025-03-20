from alpaca.data.live import CryptoDataStream


from src.settings import config


# async handler
async def quote_data_handler(data):
    """ handler for stream data """
    # quote data will arrive here
    print(data)


class CryptoDataStreamer:
    def __new__(cls, *args, **kwargs):  # , quote_data_handler: Callable):
        cds = CryptoDataStream(config['ALPACA_KEY'], config['ALPACA_SECRET'])
        cds.subscribe_daily_bars(quote_data_handler, 'BTC/USD')

    # def run(self):
    #     self.wss_client.run()
    #
    # def subscribe_bars(self, subscribe_bars_handler: Callable, ticker_symbols: list = None):
    #     self.wss_client.subscribe_bars(subscribe_bars_handler, *ticker_symbols)
    #
    # def unsubscribe_bars(self, ticker_symbols: list = None):
    #     self.wss_client.unsubscribe_bars(*ticker_symbols)

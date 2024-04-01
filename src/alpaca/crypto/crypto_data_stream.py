""" streams alpaca crypto data """

from alpaca.data.live import CryptoDataStream

from src.settings import config

wss_client = CryptoDataStream(config['ALPACA_KEY'], config['ALPACA_SECRET'])


# async handler
async def quote_data_handler(data):
    """ handler for stream data """
    # quote data will arrive here
    print(data)

wss_client.subscribe_bars(quote_data_handler, "BTC/USD")
wss_client.run()

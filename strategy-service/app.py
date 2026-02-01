
from src.trading_bot import TradingBot
from src.strategies.simple_strategy import SimpleStrategy
from src.logger import logger

assets_to_trade = ["BTC/USD", "ETH/USD"]


if __name__ == '__main__':
    trading_strategy = SimpleStrategy(assets_to_trade)
    trading_bot = TradingBot(trading_strategy)
    trading_bot.run()

    if not trading_bot.is_able_to_trade():
        logger.error('Cannot Trade')
        exit(1)

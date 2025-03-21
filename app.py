from src.trading_bot import TradingBot
from src.strategies.simple_strategy import SimpleStrategy


if __name__ == '__main__':
    trading_strategy = SimpleStrategy()
    trading_bot = TradingBot(trading_strategy)
    trading_bot.run()

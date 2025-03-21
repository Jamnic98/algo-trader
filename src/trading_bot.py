import sys

from src.alpaca.account import account
from src.utils import logger


class TradingBot:
    def __init__(self, trading_strategy):
        self.trading_strategy = trading_strategy

    @staticmethod
    def __is_able_to_trade() -> bool:
        return all((not account.account_blocked, not account.trading_blocked))

    def run(self):
        if not self.__is_able_to_trade():
            logger.error('Trading bot is unable to trade')
            sys.exit(1)

        # TODO: check that the specified tickers are currently tradeable

        self.trading_strategy.run()

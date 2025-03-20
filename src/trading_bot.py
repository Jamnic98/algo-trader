from typing import Any, Callable, Tuple, List
from alpaca.trading import Position
from src.alpaca.account import account
from src.alpaca.client import trading_client
from src.strategies.trading_strategy import TradingStrategy


class TradingBot:
    def __init__(self, trading_strategy):
        self.trading_strategy = trading_strategy

    @staticmethod
    def is_able_to_trade() -> bool:
        return all((not account.account_blocked, not account.trading_blocked))

    def run(self):
        # check that bot is able to make trades
        # check that the specified tickers are currently tradeable
        try:
            strategy = self.trading_strategy()
            strategy.run()

        except Exception as _e:
            pass

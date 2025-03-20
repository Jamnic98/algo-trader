from typing import List
from src.strategies.trading_strategy import TradingStrategy


class SimpleStrategy(TradingStrategy):
    def __init__(self, tickers: List[str]):
        super().__init__(tickers)

    async def data_handler(self):
        pass

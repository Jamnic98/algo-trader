from enum import StrEnum

ASSETS_TO_TRADE = ("BTC/USD", "ETH/USD")

class Environment(StrEnum):
    DEV = 'development'
    PROD = 'production'

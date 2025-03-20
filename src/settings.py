import os
from enum import StrEnum
from dotenv import load_dotenv
from src.logger import logger

class Environment(StrEnum):
    DEV = 'development'
    PROD = 'production'

load_dotenv()

config = {}

environment = os.getenv('ENV', Environment.DEV.value)
if environment == Environment.DEV.value:
    config['DEBUG'] = True
    config['LIVE_TRADING'] = False
    config['ALPACA_KEY'] = os.getenv('ALPACA_PAPER_API_KEY')
    config['ALPACA_SECRET'] = os.getenv('ALPACA_PAPER_API_SECRET')
    config['DATABASE_URI'] = os.getenv('DEV_DB_URI')
elif environment == Environment.PROD.value:
    config['DEBUG'] = False
    config['LIVE_TRADING'] = True
    config['ALPACA_KEY'] = os.getenv('ALPACA_LIVE_API_KEY')
    config['ALPACA_SECRET'] = os.getenv('ALPACA_LIVE_API_SECRET')
    config['DATABASE_URI'] = os.getenv('PROD_DB_URI')
else:
    raise ValueError(f'Unknown environment: {environment}')


logger.debug("Running in %s mode", environment)

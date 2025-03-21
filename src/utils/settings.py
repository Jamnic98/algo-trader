import os
import sys
from dotenv import load_dotenv

from src.utils import get_env_var
from src.utils.constants import Environment
from src.utils.logging import logger


if os.path.exists('.env'):
    load_dotenv()
else:
    logger.warning(".env file not found, environment variables may be missing.")

# Default configuration
class Config:
    def __init__(self):
        self.DEBUG = True  # pylint: disable=invalid-name
        self.LIVE_TRADING = False  # pylint: disable=invalid-name
        self.ALPACA_KEY = get_env_var('ALPACA_PAPER_API_KEY')  # pylint: disable=invalid-name
        self.ALPACA_SECRET = get_env_var('ALPACA_PAPER_API_SECRET')  # pylint: disable=invalid-name
        self.DATABASE_URI = get_env_var('DEV_DB_URI')  # pylint: disable=invalid-name

config = Config().__dict__
environment = get_env_var('ENV', required=False) or Environment.DEV.value
logger.debug("Running in %s mode with DB: %s", environment, config['DATABASE_URI'])

if environment == Environment.PROD.value:
    config.update({
        'DEBUG': False,
        'LIVE_TRADING': True,
        'ALPACA_KEY': get_env_var('ALPACA_LIVE_API_KEY'),
        'ALPACA_SECRET': get_env_var('ALPACA_LIVE_API_SECRET'),
        'DATABASE_URI': get_env_var('PROD_DB_URI')
    })
elif environment != Environment.DEV.value:
    logger.error("Invalid environment detected: %s", environment)
    sys.exit(1)

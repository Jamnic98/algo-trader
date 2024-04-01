import os
from dotenv import load_dotenv
from src.logger import logger

load_dotenv()

config = {
    'DEBUG': True,
    'LIVE_TRADING': False,
    'ALPACA_KEY': os.getenv('ALPACA_PAPER_API_KEY'),
    'ALPACA_SECRET': os.getenv('ALPACA_PAPER_API_SECRET'),
    'DATABASE_URI': os.getenv('DEV_DB_URI')
}

environment = os.getenv('ENV', 'development')
if environment == 'development':
    pass
elif environment == 'production':
    # Production environment settings
    config['DEBUG'] = False
    config['LIVE_TRADING'] = True
    config['ALPACA_KEY'] = os.getenv('ALPACA_LIVE_API_KEY')
    config['ALPACA_SECRET'] = os.getenv('ALPACA_LIVE_API_SECRET')
    config['DATABASE_URI'] = os.getenv('PROD_DB_URI')
else:
    raise ValueError(f'Unknown environment: {environment}')


logger.debug("Running in %s mode", environment)

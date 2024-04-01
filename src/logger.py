import logging
from pathlib import Path

root_path = Path(__file__).parent.parent

# Create a generic logger
logger = logging.getLogger('logger')
logger.setLevel(logging.DEBUG)

# Set a formatter for the handlers
formatter = logging.Formatter(
    '%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    datefmt='%d/%m/%y %H:%M:%S'
)

# Create a file handler for logging trade info
# file_handler = logging.FileHandler(Path.joinpath(root_path, 'trades.log'))
# file_handler.setLevel(logging.INFO)
# file_handler.setFormatter(formatter)
# logger.addHandler(file_handler)

# Create a console handler
console_handler = logging.StreamHandler()
console_handler.setLevel(logging.DEBUG)
console_handler.setFormatter(formatter)
logger.addHandler(console_handler)

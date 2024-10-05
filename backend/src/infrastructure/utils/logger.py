import logging
from src.infrastructure.logging.colored_formatter import ColoredFormatter

def setup_logger(dev_mode):
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.DEBUG)
    ch = logging.StreamHandler()
    ch.setLevel(logging.DEBUG if dev_mode.lower() == 'true' else logging.INFO)
    ch.setFormatter(ColoredFormatter())
    logger.addHandler(ch)
    return logger

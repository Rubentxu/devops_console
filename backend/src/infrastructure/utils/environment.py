import os
from dotenv import load_dotenv

def load_environment_variables():
    load_dotenv()
    return int(os.getenv("API_PORT", 8080)), os.getenv('DEV_MODE', 'False')

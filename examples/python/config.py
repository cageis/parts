# Main application configuration
import os

# Base settings
DEBUG = False
SECRET_KEY = os.environ.get('SECRET_KEY', 'default-secret')

# Database configuration
DATABASE_URL = 'sqlite:///app.db'
# ============================
# PARTIALS>>>>>
# ============================
# Cache configuration
CACHE_CONFIG = {
    'backend': 'redis',
    'location': '127.0.0.1:6379',
    'timeout': 300
}
# Logging configuration
LOGGING = {
    'version': 1,
    'handlers': {
        'console': {
            'class': 'logging.StreamHandler',
        }
    },
    'root': {
        'level': 'INFO',
        'handlers': ['console']
    }
}
# ============================
# PARTIALS<<<<<
# ============================

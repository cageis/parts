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
"""Common decorators for controllers and other layers."""

import logging
from typing import Callable
from functools import wraps
from fastapi import HTTPException

logger = logging.getLogger(__name__)


def handle_http_exceptions(func: Callable) -> Callable:
    """Decorator to handle HTTP exceptions gracefully in controllers.

    - Re-raises HTTPExceptions from service layer (400, 404, etc.)
    - Catches unexpected errors, logs them, and returns 500
    - Adds full stack trace to logs for debugging

    Usage:
        @handle_http_exceptions
        def my_controller_method(param: str) -> dict:
            result = my_service_method(param)
            return result
    """
    @wraps(func)
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except HTTPException:
            raise  # Pass through expected HTTP errors
        except Exception as e:
            logger.error(f"Unexpected error in {func.__name__}: {e}", exc_info=True)
            raise HTTPException(status_code=500, detail="Internal server error")
    return wrapper

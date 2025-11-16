"""Error logging utilities for API routes."""

import logging
import traceback
from functools import wraps
from typing import Callable
from fastapi import Request, HTTPException

logger = logging.getLogger(__name__)


def log_errors(func: Callable):
    """
    Decorator to log detailed error information for route handlers.

    Logs:
    - Request method, path, headers, and query params
    - Error status code and detail
    - Full traceback for non-HTTPException errors
    """

    @wraps(func)
    async def wrapper(request: Request, *args, **kwargs):
        try:
            return await func(request, *args, **kwargs)
        except HTTPException as e:
            logger.error(
                f"HTTP {e.status_code} | {request.method} {request.url.path}\n"
                f"  Query: {dict(request.query_params)}\n"
                f"  Headers: {dict(request.headers)}\n"
                f"  Error: {e.detail}"
            )
            raise
        except Exception as e:
            logger.error(
                f"UNHANDLED ERROR | {request.method} {request.url.path}\n"
                f"  Query: {dict(request.query_params)}\n"
                f"  Error: {type(e).__name__}: {str(e)}\n"
                f"{traceback.format_exc()}"
            )
            raise

    return wrapper

"""FastAPI middleware for authentication and error logging."""

import logging
import traceback
from fastapi import Request, HTTPException
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

from lib.auth import _extract_token, verify_token, _get_email_from_claims, _get_tenant_id
from lib.audit_context import set_current_user_id
from services.auth_service import get_or_create_user

logger = logging.getLogger(__name__)

# Paths that don't require authentication
PUBLIC_PATHS = frozenset({
    "/",
    "/health",
    "/version",
    "/docs",
    "/redoc",
    "/openapi.json",
})

# Path prefixes that don't require authentication
PUBLIC_PREFIXES = (
    "/ws",  # WebSocket has its own auth
)


class AuthMiddleware(BaseHTTPMiddleware):
    """Middleware to protect routes with JWT authentication."""

    async def dispatch(self, request: Request, call_next):
        path = request.url.path

        # Skip auth for CORS preflight requests
        if request.method == "OPTIONS":
            return await call_next(request)

        # Skip auth for public paths
        if path in PUBLIC_PATHS:
            return await call_next(request)

        # Skip auth for public prefixes
        for prefix in PUBLIC_PREFIXES:
            if path.startswith(prefix):
                return await call_next(request)

        # Authenticate
        try:
            token = _extract_token(request.headers.get("Authorization"))
            claims = verify_token(token)

            request.state.auth0_sub = claims.get("sub")
            request.state.email = _get_email_from_claims(claims)

            user = await get_or_create_user(claims.get("sub"), request.state.email)
            request.state.user = user
            request.state.user_id = user.id
            request.state.tenant_id = _get_tenant_id(request, claims, user)

            # Set user_id in context for automatic audit column population
            set_current_user_id(user.id)

        except HTTPException as e:
            return JSONResponse(
                status_code=e.status_code,
                content={"detail": e.detail},
            )

        return await call_next(request)


class ErrorLoggingMiddleware(BaseHTTPMiddleware):
    """Middleware to log detailed error information."""

    async def dispatch(self, request: Request, call_next):
        try:
            return await call_next(request)
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

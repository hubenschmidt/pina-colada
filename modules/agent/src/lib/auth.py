"""JWT authentication and authorization for FastAPI."""

import os
import logging
from functools import wraps
from typing import Optional, Callable
from fastapi import HTTPException, Request
from jose import jwt, JWTError
import requests

from services.auth_service import get_or_create_user

logger = logging.getLogger(__name__)
# Cache for JWKS
_jwks_cache: Optional[dict] = None


def get_jwks() -> dict:
    """Fetch Auth0 public keys for JWT verification."""
    global _jwks_cache
    if _jwks_cache:
        return _jwks_cache

    auth0_domain = os.getenv("AUTH0_DOMAIN")
    if not auth0_domain:
        raise ValueError("AUTH0_DOMAIN environment variable not set")

    jwks_url = f"https://{auth0_domain}/.well-known/jwks.json"
    response = requests.get(jwks_url)
    response.raise_for_status()
    _jwks_cache = response.json()
    return _jwks_cache


def _find_rsa_key(jwks: dict, kid: str) -> Optional[dict]:
    """Find RSA key by kid from JWKS."""
    for key in jwks["keys"]:
        if key["kid"] == kid:
            return {
                "kty": key["kty"],
                "kid": key["kid"],
                "use": key["use"],
                "n": key["n"],
                "e": key["e"],
            }
    return None


def verify_token(token: str) -> dict:
    """Verify JWT token and return claims."""
    auth0_domain = os.getenv("AUTH0_DOMAIN")
    auth0_audience = os.getenv("AUTH0_AUDIENCE")

    if not auth0_domain or not auth0_audience:
        raise HTTPException(
            status_code=500, detail="AUTH0_DOMAIN or AUTH0_AUDIENCE not configured"
        )

    try:
        unverified_header = jwt.get_unverified_header(token)
        jwks = get_jwks()
        rsa_key = _find_rsa_key(jwks, unverified_header["kid"])

        if not rsa_key:
            raise HTTPException(
                status_code=401, detail="Unable to find appropriate key"
            )

        payload = jwt.decode(
            token,
            rsa_key,
            algorithms=["RS256"],
            audience=auth0_audience,
            issuer=f"https://{auth0_domain}/",
        )
        return payload

    except JWTError as e:
        raise HTTPException(status_code=401, detail=f"Invalid token: {str(e)}")


def _extract_token(auth_header: Optional[str]) -> str:
    """Extract and validate token from Authorization header."""
    if not auth_header:
        raise HTTPException(status_code=401, detail="Missing Authorization header")

    if not auth_header.startswith("Bearer "):
        raise HTTPException(
            status_code=401, detail="Invalid Authorization header format"
        )

    token = auth_header.split(" ")[1].strip()
    if not token:
        raise HTTPException(status_code=401, detail="Token is empty")

    token_parts = token.split(".")
    if len(token_parts) != 3:
        detail = "Unauthorized"
        if os.getenv("ENVIRONMENT") == "development":
            detail = f"Invalid token format. Expected JWT format (3 parts), got {len(token_parts)} parts"
        raise HTTPException(status_code=401, detail=detail)

    return token


def _get_email_from_claims(claims: dict) -> Optional[str]:
    """Extract email from standard or namespaced claim."""
    email = claims.get("email")
    if email:
        return email

    auth0_domain = os.getenv("AUTH0_DOMAIN")
    if not auth0_domain:
        return None

    return claims.get(f"https://{auth0_domain}/email")


def _get_tenant_id(request: Request, claims: dict, user) -> Optional[int]:
    """Get tenant ID from header, claim, or user."""
    tenant_id = request.headers.get("X-Tenant-Id")
    if tenant_id:
        return int(tenant_id)

    namespace = os.getenv("AUTH0_NAMESPACE", "https://pinacolada.co")
    claim_tenant = claims.get(f"{namespace}/tenant_id")
    if claim_tenant:
        return int(claim_tenant)

    return user.tenant_id if user.tenant_id else None


def require_auth(func: Callable):
    """Decorator to protect routes with JWT authentication."""

    @wraps(func)
    async def wrapper(request: Request, *args, **kwargs):
        token = _extract_token(request.headers.get("Authorization"))
        claims = verify_token(token)

        request.state.auth0_sub = claims.get("sub")
        request.state.email = _get_email_from_claims(claims)

        user = await get_or_create_user(claims.get("sub"), request.state.email)
        request.state.user = user
        request.state.user_id = user.id
        request.state.tenant_id = _get_tenant_id(request, claims, user)

        return await func(request, *args, **kwargs)

    return wrapper

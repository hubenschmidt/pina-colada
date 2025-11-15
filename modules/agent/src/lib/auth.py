"""JWT authentication and authorization for FastAPI."""

import os
from functools import wraps
from typing import Optional, Callable
from fastapi import HTTPException, Request
from jose import jwt, JWTError
import requests


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


def verify_token(token: str) -> dict:
    """Verify JWT token and return claims."""
    auth0_domain = os.getenv("AUTH0_DOMAIN")
    auth0_audience = os.getenv("AUTH0_AUDIENCE")

    if not auth0_domain or not auth0_audience:
        raise HTTPException(
            status_code=500,
            detail="AUTH0_DOMAIN or AUTH0_AUDIENCE not configured"
        )

    try:
        unverified_header = jwt.get_unverified_header(token)
        jwks = get_jwks()
        rsa_key = {}

        for key in jwks["keys"]:
            if key["kid"] == unverified_header["kid"]:
                rsa_key = {
                    "kty": key["kty"],
                    "kid": key["kid"],
                    "use": key["use"],
                    "n": key["n"],
                    "e": key["e"]
                }
                break

        if not rsa_key:
            raise HTTPException(
                status_code=401,
                detail="Unable to find appropriate key"
            )

        payload = jwt.decode(
            token,
            rsa_key,
            algorithms=["RS256"],
            audience=auth0_audience,
            issuer=f"https://{auth0_domain}/"
        )
        return payload

    except JWTError as e:
        raise HTTPException(
            status_code=401,
            detail=f"Invalid token: {str(e)}"
        )


def require_auth(func: Callable):
    """Decorator to protect routes with JWT authentication."""
    @wraps(func)
    async def wrapper(request: Request, *args, **kwargs):
        auth_header = request.headers.get("Authorization")
        if not auth_header:
            raise HTTPException(
                status_code=401,
                detail="Missing Authorization header"
            )

        if not auth_header.startswith("Bearer "):
            raise HTTPException(
                status_code=401,
                detail="Invalid Authorization header format"
            )

        token = auth_header.split(" ")[1].strip()
        if not token:
            raise HTTPException(
                status_code=401,
                detail="Token is empty"
            )
        
        # Basic JWT format validation (should have 3 parts separated by dots)
        if len(token.split(".")) != 3:
            raise HTTPException(
                status_code=401,
                detail=f"Invalid token format. Expected JWT format (3 parts), got {len(token.split('.'))} parts"
            )
        
        claims = verify_token(token)

        # Attach user info to request state
        request.state.auth0_sub = claims.get("sub")
        
        # Get email from standard claim or namespaced custom claim (Auth0 action adds it)
        email = claims.get("email")
        if not email:
            # Check namespaced custom claim (Auth0 action namespace)
            auth0_domain = os.getenv("AUTH0_DOMAIN")
            if auth0_domain:
                email = claims.get(f"https://{auth0_domain}/email")
        
        request.state.email = email

        # Load user from database
        from services.auth_service import get_or_create_user
        user = await get_or_create_user(claims.get("sub"), email)
        request.state.user = user
        request.state.user_id = user.id

        # Get tenant from header or custom claim
        tenant_id = request.headers.get("X-Tenant-Id")
        if not tenant_id:
            # Try custom claim
            namespace = "https://pinacolada.co"
            tenant_id = claims.get(f"{namespace}/tenant_id")

        if tenant_id:
            request.state.tenant_id = int(tenant_id)
        elif user.tenant_id:
            # Default to user's primary tenant
            request.state.tenant_id = user.tenant_id
        else:
            request.state.tenant_id = None

        return await func(request, *args, **kwargs)

    return wrapper

"""Routes for accounts API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.account_controller import search_accounts


router = APIRouter(prefix="/accounts", tags=["accounts"])


@router.get("/search")
@log_errors
@require_auth
async def search_accounts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search accounts by name."""
    if not q:
        return []
    return await search_accounts(request, q)

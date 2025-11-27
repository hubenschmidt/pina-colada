"""Contact API routes."""

from typing import Optional
from fastapi import APIRouter, Request, Query
from lib.error_logging import log_errors
from lib.auth import require_auth
from repositories.contact_repository import search_contacts_and_individuals

router = APIRouter(prefix="/contacts", tags=["contacts"])


@router.get("/search")
@log_errors
@require_auth
async def search_contacts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """
    Search for contacts/individuals by name or email.
    Returns individuals that can be linked as contacts.
    """
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    results = await search_contacts_and_individuals(q, tenant_id=tenant_id)
    return results

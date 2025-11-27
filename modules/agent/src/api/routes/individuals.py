"""Routes for individuals API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request, Query
from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.individual_repository import find_all_individuals, search_individuals

router = APIRouter(prefix="/individuals", tags=["individuals"])


def _ind_to_dict(ind):
    return {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
        "phone": ind.phone,
        "linkedin_url": ind.linkedin_url,
        "title": ind.title,
        "notes": ind.notes,
        "created_at": ind.created_at.isoformat() if ind.created_at else None,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }


@router.get("")
@log_errors
@require_auth
async def get_individuals_route(request: Request):
    """Get all individuals for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    individuals = await find_all_individuals(tenant_id=tenant_id)
    return [_ind_to_dict(ind) for ind in individuals]


@router.get("/search")
@log_errors
@require_auth
async def search_individuals_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search individuals by name or email."""
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    individuals = await search_individuals(q, tenant_id=tenant_id)
    return [_ind_to_dict(ind) for ind in individuals]

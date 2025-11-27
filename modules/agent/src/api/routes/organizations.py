"""Routes for organizations API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request, Query
from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.organization_repository import find_all_organizations, search_organizations

router = APIRouter(prefix="/organizations", tags=["organizations"])


def _org_to_dict(org):
    return {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "phone": org.phone,
        "industries": [ind.name for ind in org.industries] if org.industries else [],
        "employee_count": org.employee_count,
        "description": org.description,
        "created_at": org.created_at.isoformat() if org.created_at else None,
        "updated_at": org.updated_at.isoformat() if org.updated_at else None,
    }


@router.get("")
@log_errors
@require_auth
async def get_organizations_route(request: Request):
    """Get all organizations for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await find_all_organizations(tenant_id=tenant_id)
    return [_org_to_dict(org) for org in organizations]


@router.get("/search")
@log_errors
@require_auth
async def search_organizations_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search organizations by name."""
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await search_organizations(q, tenant_id=tenant_id)
    return [_org_to_dict(org) for org in organizations]

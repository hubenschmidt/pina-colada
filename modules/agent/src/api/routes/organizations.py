"""Routes for organizations API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.organization_repository import find_all_organizations

router = APIRouter(prefix="/organizations", tags=["organizations"])


@router.get("")
@log_errors
@require_auth
async def get_organizations_route(request: Request):
    """Get all organizations for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await find_all_organizations(tenant_id=tenant_id)
    return [
        {
            "id": org.id,
            "name": org.name,
            "website": org.website,
            "phone": org.phone,
            "industries": [ind.name for ind in org.industries],
            "employee_count": org.employee_count,
            "description": org.description,
            "created_at": org.created_at.isoformat() if org.created_at else None,
            "updated_at": org.updated_at.isoformat() if org.updated_at else None,
        }
        for org in organizations
    ]

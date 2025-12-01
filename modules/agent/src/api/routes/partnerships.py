"""Routes for partnerships API endpoints."""

from typing import Optional, List, Type
from fastapi import APIRouter, Query, Request
from pydantic import BaseModel, create_model
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.partnership_controller import (
    get_partnerships,
    create_partnership,
    get_partnership,
    update_partnership,
    delete_partnership,
)

router = APIRouter(prefix="/partnerships", tags=["partnerships"])


def _make_partnership_create_model() -> Type[BaseModel]:
    """Create PartnershipCreate model functionally."""
    return create_model(
        "PartnershipCreate",
        account_type=(str, "Organization"),
        account=(Optional[str], None),
        contacts=(Optional[List[dict]], None),
        industry=(Optional[List[str]], None),
        industry_ids=(Optional[List[int]], None),
        title=(str, ...),
        partnership_name=(str, ...),
        partnership_type=(Optional[str], None),
        start_date=(Optional[str], None),
        end_date=(Optional[str], None),
        description=(Optional[str], None),
        status=(str, "Exploring"),
        source=(str, "manual"),
        project_ids=(Optional[List[int]], None),
    )


def _make_partnership_update_model() -> Type[BaseModel]:
    """Create PartnershipUpdate model functionally."""
    return create_model(
        "PartnershipUpdate",
        account=(Optional[str], None),
        contacts=(Optional[List[dict]], None),
        title=(Optional[str], None),
        partnership_name=(Optional[str], None),
        partnership_type=(Optional[str], None),
        start_date=(Optional[str], None),
        end_date=(Optional[str], None),
        description=(Optional[str], None),
        status=(Optional[str], None),
        source=(Optional[str], None),
        project_ids=(Optional[List[int]], None),
    )


PartnershipCreate = _make_partnership_create_model()
PartnershipUpdate = _make_partnership_update_model()


@router.get("")
@log_errors
@require_auth
async def get_partnerships_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
    project_id: Optional[int] = Query(None, alias="projectId"),
):
    """Get all partnerships with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    return await get_partnerships(page, limit, order_by, order, search, tenant_id, project_id)


@router.post("")
@log_errors
@require_auth
async def create_partnership_route(request: Request, data: PartnershipCreate):
    """Create a new partnership."""
    partner_data = data.dict()
    partner_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    return await create_partnership(partner_data)


@router.get("/{partnership_id}")
@log_errors
@require_auth
async def get_partnership_route(request: Request, partnership_id: str):
    """Get a partnership by ID."""
    return await get_partnership(partnership_id)


@router.put("/{partnership_id}")
@log_errors
@require_auth
async def update_partnership_route(request: Request, partnership_id: str, data: PartnershipUpdate):
    """Update a partnership."""
    return await update_partnership(partnership_id, data.dict(exclude_unset=True))


@router.delete("/{partnership_id}")
@log_errors
@require_auth
async def delete_partnership_route(request: Request, partnership_id: str):
    """Delete a partnership."""
    return await delete_partnership(partnership_id)

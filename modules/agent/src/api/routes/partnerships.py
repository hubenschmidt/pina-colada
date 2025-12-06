"""Routes for partnerships API endpoints."""

from typing import Optional

from fastapi import APIRouter, Query, Request

from controllers.partnership_controller import (
    create_partnership,
    delete_partnership,
    get_partnership,
    get_partnerships,
    update_partnership,
)
from schemas.partnership import PartnershipCreate, PartnershipUpdate


router = APIRouter(prefix="/partnerships", tags=["partnerships"])


@router.get("")
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
    return await get_partnerships(request, page, limit, order_by, order, search, project_id)


@router.post("")
async def create_partnership_route(request: Request, data: PartnershipCreate):
    """Create a new partnership."""
    return await create_partnership(request, data)


@router.get("/{partnership_id}")
async def get_partnership_route(request: Request, partnership_id: str):
    """Get a partnership by ID."""
    return await get_partnership(partnership_id)


@router.put("/{partnership_id}")
async def update_partnership_route(request: Request, partnership_id: str, data: PartnershipUpdate):
    """Update a partnership."""
    return await update_partnership(request, partnership_id, data)


@router.delete("/{partnership_id}")
async def delete_partnership_route(request: Request, partnership_id: str):
    """Delete a partnership."""
    return await delete_partnership(partnership_id)

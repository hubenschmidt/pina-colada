"""Routes for opportunities API endpoints."""

from typing import Optional

from fastapi import APIRouter, Query, Request

from controllers.opportunity_controller import (
    get_opportunities,
    create_opportunity,
    get_opportunity,
    update_opportunity,
    delete_opportunity,
    OpportunityCreate,
    OpportunityUpdate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/opportunities", tags=["opportunities"])


@router.get("")
@log_errors
@require_auth
async def get_opportunities_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
    project_id: Optional[int] = Query(None, alias="projectId"),
):
    """Get all opportunities with pagination."""
    return await get_opportunities(request, page, limit, order_by, order, search, project_id)


@router.post("")
@log_errors
@require_auth
async def create_opportunity_route(request: Request, data: OpportunityCreate):
    """Create a new opportunity."""
    return await create_opportunity(request, data)


@router.get("/{opportunity_id}")
@log_errors
@require_auth
async def get_opportunity_route(request: Request, opportunity_id: str):
    """Get an opportunity by ID."""
    return await get_opportunity(opportunity_id)


@router.put("/{opportunity_id}")
@log_errors
@require_auth
async def update_opportunity_route(request: Request, opportunity_id: str, data: OpportunityUpdate):
    """Update an opportunity."""
    return await update_opportunity(request, opportunity_id, data)


@router.delete("/{opportunity_id}")
@log_errors
@require_auth
async def delete_opportunity_route(request: Request, opportunity_id: str):
    """Delete an opportunity."""
    return await delete_opportunity(opportunity_id)

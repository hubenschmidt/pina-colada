"""Routes for opportunities API endpoints."""

from typing import Optional

from fastapi import APIRouter, Query, Request

from controllers.opportunity_controller import (
    create_opportunity,
    delete_opportunity,
    get_opportunities,
    get_opportunity,
    update_opportunity,
)
from schemas.opportunity import OpportunityCreate, OpportunityUpdate


router = APIRouter(prefix="/opportunities", tags=["opportunities"])


@router.get("")
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
async def create_opportunity_route(request: Request, data: OpportunityCreate):
    """Create a new opportunity."""
    return await create_opportunity(request, data)


@router.get("/{opportunity_id}")
async def get_opportunity_route(request: Request, opportunity_id: str):
    """Get an opportunity by ID."""
    return await get_opportunity(opportunity_id)


@router.put("/{opportunity_id}")
async def update_opportunity_route(request: Request, opportunity_id: str, data: OpportunityUpdate):
    """Update an opportunity."""
    return await update_opportunity(request, opportunity_id, data)


@router.delete("/{opportunity_id}")
async def delete_opportunity_route(request: Request, opportunity_id: str):
    """Delete an opportunity."""
    return await delete_opportunity(opportunity_id)

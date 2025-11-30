"""Routes for opportunities API endpoints."""

from typing import Optional, List
from fastapi import APIRouter, Query, Request
from pydantic import BaseModel, Field
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.opportunity_controller import (
    get_opportunities,
    create_opportunity,
    get_opportunity,
    update_opportunity,
    delete_opportunity,
)

router = APIRouter(prefix="/opportunities", tags=["opportunities"])


class OpportunityCreate(BaseModel):
    """Model for creating an opportunity."""
    account_type: str = "Organization"
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    industry: Optional[List[str]] = None
    industry_ids: Optional[List[int]] = None
    title: str
    opportunity_name: str
    estimated_value: Optional[float] = None
    probability: Optional[int] = Field(default=None, ge=0, le=100)
    expected_close_date: Optional[str] = None
    notes: Optional[str] = None
    status: str = "Qualifying"
    source: str = "manual"
    project_ids: Optional[List[int]] = None


class OpportunityUpdate(BaseModel):
    """Model for updating an opportunity."""
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    title: Optional[str] = None
    opportunity_name: Optional[str] = None
    estimated_value: Optional[float] = None
    probability: Optional[int] = Field(default=None, ge=0, le=100)
    expected_close_date: Optional[str] = None
    notes: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None
    project_ids: Optional[List[int]] = None


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
    tenant_id = getattr(request.state, "tenant_id", None)
    return await get_opportunities(page, limit, order_by, order, search, tenant_id, project_id)


@router.post("")
@log_errors
@require_auth
async def create_opportunity_route(request: Request, data: OpportunityCreate):
    """Create a new opportunity."""
    opp_data = data.dict()
    opp_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    return await create_opportunity(opp_data)


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
    return await update_opportunity(opportunity_id, data.dict(exclude_unset=True))


@router.delete("/{opportunity_id}")
@log_errors
@require_auth
async def delete_opportunity_route(request: Request, opportunity_id: str):
    """Delete an opportunity."""
    return await delete_opportunity(opportunity_id)

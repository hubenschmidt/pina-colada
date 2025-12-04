"""Controller layer for opportunity routing to services."""

from typing import Optional, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response
from serializers.opportunity import (
    opportunity_to_list_response,
    opportunity_to_detail_response,
)
from schemas.opportunity import OpportunityCreate, OpportunityUpdate
from services.opportunity_service import (
    create_opportunity as create_opportunity_service,
    delete_opportunity as delete_opportunity_service,
    get_opportunities_paginated,
    get_opportunity as get_opportunity_service,
    update_opportunity as update_opportunity_service,
)


@handle_http_exceptions
async def get_opportunities(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
    project_id: Optional[int] = None,
) -> dict:
    """Get all opportunities with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    paginated, total_count = await get_opportunities_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [opportunity_to_list_response(opp) for opp in paginated]
    return to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_opportunity(request: Request, data: OpportunityCreate) -> Dict[str, Any]:
    """Create a new opportunity."""
    opp_data = data.dict()
    opp_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    opp_data["user_id"] = getattr(request.state, "user_id", None)
    created = await create_opportunity_service(opp_data)
    return opportunity_to_detail_response(created)


@handle_http_exceptions
async def get_opportunity(opp_id: str) -> Dict[str, Any]:
    """Get an opportunity by ID."""
    opp = await get_opportunity_service(opp_id)
    return opportunity_to_detail_response(opp)


@handle_http_exceptions
async def update_opportunity(request: Request, opp_id: str, data: OpportunityUpdate) -> Dict[str, Any]:
    """Update an opportunity."""
    update_data = data.dict(exclude_unset=True)
    update_data["user_id"] = getattr(request.state, "user_id", None)
    updated = await update_opportunity_service(opp_id, update_data)
    return opportunity_to_detail_response(updated)


@handle_http_exceptions
async def delete_opportunity(opp_id: str) -> dict:
    """Delete an opportunity."""
    await delete_opportunity_service(opp_id)
    return {"success": True}

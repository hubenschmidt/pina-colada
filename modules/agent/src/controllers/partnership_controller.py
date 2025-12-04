"""Controller layer for partnership routing to services."""

from typing import Optional, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response
from serializers.partnership import (
    partnership_to_list_response,
    partnership_to_detail_response,
)
from services.partnership_service import (
    PartnershipCreate,
    PartnershipUpdate,
    get_partnerships_paginated,
    create_partnership as create_partnership_service,
    get_partnership as get_partnership_service,
    update_partnership as update_partnership_service,
    delete_partnership as delete_partnership_service,
)

# Re-export for routes
__all__ = ["PartnershipCreate", "PartnershipUpdate"]


@handle_http_exceptions
async def get_partnerships(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
    project_id: Optional[int] = None,
) -> dict:
    """Get all partnerships with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    paginated, total_count = await get_partnerships_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [partnership_to_list_response(p) for p in paginated]
    return to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_partnership(request: Request, data: PartnershipCreate) -> Dict[str, Any]:
    """Create a new partnership."""
    partner_data = data.dict()
    partner_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    partner_data["user_id"] = getattr(request.state, "user_id", None)
    created = await create_partnership_service(partner_data)
    return partnership_to_detail_response(created)


@handle_http_exceptions
async def get_partnership(partnership_id: str) -> Dict[str, Any]:
    """Get a partnership by ID."""
    partnership = await get_partnership_service(partnership_id)
    return partnership_to_detail_response(partnership)


@handle_http_exceptions
async def update_partnership(request: Request, partnership_id: str, data: PartnershipUpdate) -> Dict[str, Any]:
    """Update a partnership."""
    update_data = data.dict(exclude_unset=True)
    update_data["user_id"] = getattr(request.state, "user_id", None)
    updated = await update_partnership_service(partnership_id, update_data)
    return partnership_to_detail_response(updated)


@handle_http_exceptions
async def delete_partnership(partnership_id: str) -> dict:
    """Delete a partnership."""
    await delete_partnership_service(partnership_id)
    return {"success": True}

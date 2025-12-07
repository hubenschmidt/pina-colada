"""Controller layer for usage analytics routing to services."""

from typing import List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from lib.role_check import user_has_role
from services import usage_service as service


@handle_http_exceptions
async def get_user_usage(request: Request, period: str = "monthly") -> dict:
    """Get aggregated usage for the current user."""
    user_id = request.state.user_id
    return await service.get_user_usage(user_id, period)


@handle_http_exceptions
async def get_tenant_usage(request: Request, period: str = "monthly") -> dict:
    """Get aggregated usage for the tenant."""
    tenant_id = request.state.tenant_id
    return await service.get_tenant_usage(tenant_id, period)


@handle_http_exceptions
async def get_usage_timeseries(
    request: Request,
    period: str = "monthly",
    scope: str = "user",
) -> List[dict]:
    """Get usage timeseries data for charts."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id if scope == "user" else None
    return await service.get_usage_timeseries(tenant_id, period, user_id)


@handle_http_exceptions
async def get_developer_analytics(
    request: Request,
    period: str = "monthly",
    group_by: str = "node",
) -> dict:
    """Get developer analytics breakdown (requires developer role)."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id

    has_role = await user_has_role(user_id, "developer")
    if not has_role:
        return {"error": "Insufficient permissions", "data": []}

    group_by_handlers = {
        "model": service.get_usage_by_model,
        "node": service.get_usage_by_node,
    }
    handler = group_by_handlers.get(group_by, service.get_usage_by_node)
    data = await handler(tenant_id, period)

    return {"group_by": group_by, "data": data}


@handle_http_exceptions
async def check_developer_access(request: Request) -> dict:
    """Check if user has developer role."""
    user_id = request.state.user_id
    has_role = await user_has_role(user_id, "developer")
    return {"has_developer_access": has_role}

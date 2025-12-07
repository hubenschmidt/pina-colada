"""Routes for usage analytics API endpoints."""

from fastapi import APIRouter, Request

from controllers.usage_controller import (
    get_user_usage,
    get_tenant_usage,
    get_usage_timeseries,
    get_developer_analytics,
    check_developer_access,
)


router = APIRouter(prefix="/usage", tags=["usage"])


@router.get("/user")
async def get_user_usage_route(request: Request, period: str = "monthly"):
    """Get aggregated usage for the current user."""
    return await get_user_usage(request, period)


@router.get("/tenant")
async def get_tenant_usage_route(request: Request, period: str = "monthly"):
    """Get aggregated usage for the tenant."""
    return await get_tenant_usage(request, period)


@router.get("/timeseries")
async def get_usage_timeseries_route(
    request: Request,
    period: str = "monthly",
    scope: str = "user",
):
    """Get usage timeseries data for charts."""
    return await get_usage_timeseries(request, period, scope)


@router.get("/analytics")
async def get_developer_analytics_route(
    request: Request,
    period: str = "monthly",
    group_by: str = "node",
):
    """Get developer analytics breakdown (requires developer role)."""
    return await get_developer_analytics(request, period, group_by)


@router.get("/developer-access")
async def check_developer_access_route(request: Request):
    """Check if user has developer role."""
    return await check_developer_access(request)

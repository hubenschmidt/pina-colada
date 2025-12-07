"""Controller for provider costs endpoints."""

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services import provider_costs_service


@handle_http_exceptions
async def get_costs_summary(request: Request, period: str = "monthly"):
    """Get combined costs from all providers (filtered to app API key)."""
    return await provider_costs_service.get_combined_costs(period)


@handle_http_exceptions
async def get_org_costs_summary(request: Request, period: str = "monthly"):
    """Get org-wide costs including Claude Code."""
    return await provider_costs_service.get_org_costs(period)

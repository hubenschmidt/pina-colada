"""Controller for provider costs endpoints."""

from fastapi import Request

from services import provider_costs_service


async def get_costs_summary(request: Request, period: str = "monthly"):
    """Get combined costs from all providers."""
    return await provider_costs_service.get_combined_costs(period)

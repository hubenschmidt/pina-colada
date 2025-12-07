"""Routes for provider costs."""

from fastapi import APIRouter, Query, Request

from controllers.costs_controller import get_costs_summary, get_org_costs_summary

router = APIRouter(prefix="/costs", tags=["costs"])


@router.get("/summary")
async def costs_summary_route(
    request: Request,
    period: str = Query(default="monthly", description="Time period for costs"),
):
    """Get combined costs from all LLM providers (filtered to app API key)."""
    return await get_costs_summary(request, period)


@router.get("/org")
async def org_costs_route(
    request: Request,
    period: str = Query(default="monthly", description="Time period for costs"),
):
    """Get org-wide costs including Claude Code usage."""
    return await get_org_costs_summary(request, period)

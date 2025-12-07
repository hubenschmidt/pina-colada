"""Routes for provider costs."""

from fastapi import APIRouter, Query, Request

from controllers.costs_controller import get_costs_summary

router = APIRouter(prefix="/costs", tags=["costs"])


@router.get("/summary")
async def costs_summary_route(
    request: Request,
    period: str = Query(default="monthly", description="Time period for costs"),
):
    """Get combined costs from all LLM providers."""
    return await get_costs_summary(request, period)

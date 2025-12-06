"""Routes for funding stages API endpoints."""

from fastapi import APIRouter, Request

from controllers.funding_stage_controller import get_all_funding_stages


router = APIRouter(prefix="/funding-stages", tags=["funding-stages"])


@router.get("")
async def get_funding_stages(request: Request):
    """Get all funding stages."""
    return await get_all_funding_stages()

"""Routes for industries API endpoints."""

from fastapi import APIRouter, Request

from controllers.industry_controller import create_industry, get_all_industries
from schemas.industry import IndustryCreate


router = APIRouter(prefix="/industries", tags=["industries"])


@router.get("")
async def get_industries(request: Request):
    """Get all industries."""
    return await get_all_industries()


@router.post("")
async def create_industry_route(request: Request, data: IndustryCreate):
    """Create a new industry."""
    return await create_industry(data)

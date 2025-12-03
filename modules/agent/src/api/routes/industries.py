"""Routes for industries API endpoints."""

from fastapi import APIRouter, Request

from controllers.industry_controller import get_all_industries, create_industry, IndustryCreate
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/industries", tags=["industries"])


@router.get("")
@log_errors
@require_auth
async def get_industries(request: Request):
    """Get all industries."""
    return await get_all_industries()


@router.post("")
@log_errors
@require_auth
async def create_industry_route(request: Request, data: IndustryCreate):
    """Create a new industry."""
    return await create_industry(data)

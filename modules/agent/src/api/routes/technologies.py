"""Routes for technologies API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.technology_controller import (
    get_all_technologies,
    get_technology,
    create_technology,
    TechnologyCreate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/technologies", tags=["technologies"])


@router.get("")
@log_errors
@require_auth
async def get_technologies_route(request: Request, category: Optional[str] = Query(None)):
    """Get all technologies, optionally filtered by category."""
    return await get_all_technologies(category=category)


@router.get("/{tech_id}")
@log_errors
@require_auth
async def get_technology_route(request: Request, tech_id: int):
    """Get a single technology by ID."""
    return await get_technology(tech_id)


@router.post("")
@log_errors
@require_auth
async def create_technology_route(request: Request, data: TechnologyCreate):
    """Create a new technology."""
    return await create_technology(data)

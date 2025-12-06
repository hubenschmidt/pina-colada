"""Routes for technologies API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.technology_controller import (
    create_technology,
    get_all_technologies,
    get_technology,
)
from schemas.technology import TechnologyCreate


router = APIRouter(prefix="/technologies", tags=["technologies"])


@router.get("")
async def get_technologies_route(request: Request, category: Optional[str] = Query(None)):
    """Get all technologies, optionally filtered by category."""
    return await get_all_technologies(category=category)


@router.get("/{tech_id}")
async def get_technology_route(request: Request, tech_id: int):
    """Get a single technology by ID."""
    return await get_technology(tech_id)


@router.post("")
async def create_technology_route(request: Request, data: TechnologyCreate):
    """Create a new technology."""
    return await create_technology(data)

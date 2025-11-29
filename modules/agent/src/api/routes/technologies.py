"""Routes for technologies API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.technology_repository import (
    find_all_technologies,
    find_technology_by_id,
    create_technology,
)

router = APIRouter(prefix="/technologies", tags=["technologies"])


class TechnologyCreate(BaseModel):
    name: str
    category: str
    vendor: Optional[str] = None


@router.get("")
@log_errors
@require_auth
async def get_technologies(request: Request, category: Optional[str] = Query(None)):
    """Get all technologies, optionally filtered by category."""
    technologies = await find_all_technologies(category=category)
    return [model_to_dict(t, include_relationships=False) for t in technologies]


@router.get("/{tech_id}")
@log_errors
@require_auth
async def get_technology(request: Request, tech_id: int):
    """Get a single technology by ID."""
    tech = await find_technology_by_id(tech_id)
    if not tech:
        raise HTTPException(status_code=404, detail="Technology not found")
    return model_to_dict(tech, include_relationships=False)


@router.post("")
@log_errors
@require_auth
async def create_technology_route(request: Request, data: TechnologyCreate):
    """Create a new technology."""
    tech = await create_technology(
        name=data.name,
        category=data.category,
        vendor=data.vendor
    )
    return model_to_dict(tech, include_relationships=False)

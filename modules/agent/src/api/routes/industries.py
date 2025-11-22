"""Routes for industries API endpoints."""

from fastapi import APIRouter, Request
from pydantic import BaseModel
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.industry_repository import find_all_industries, find_industry_by_name
from models.Industry import Industry
from lib.db import async_get_session

router = APIRouter(prefix="/industries", tags=["industries"])


class IndustryCreate(BaseModel):
    name: str


def normalize_industry_name(name: str) -> str:
    """Normalize industry name to Title Case."""
    return name.strip().title()


@router.get("")
@log_errors
@require_auth
async def get_industries(request: Request):
    """Get all industries."""
    industries = await find_all_industries()
    return [model_to_dict(ind, include_relationships=False) for ind in industries]


@router.post("")
@log_errors
@require_auth
async def create_industry(request: Request, data: IndustryCreate):
    """Create a new industry with normalized name."""
    normalized_name = normalize_industry_name(data.name)

    existing = await find_industry_by_name(normalized_name)
    if existing:
        return model_to_dict(existing, include_relationships=False)

    async with async_get_session() as session:
        industry = Industry(name=normalized_name)
        session.add(industry)
        await session.commit()
        await session.refresh(industry)
        return model_to_dict(industry, include_relationships=False)

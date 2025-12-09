"""Service layer for industry business logic."""

from typing import List

from repositories.industry_repository import (
    find_all_industries,
    find_industry_by_name,
    create_industry as create_industry_repo,
    IndustryCreate,
)

# Re-export Pydantic models for controllers
__all__ = ["IndustryCreate"]


def normalize_industry_name(name: str) -> str:
    """Normalize industry name to Title Case."""
    return name.strip().title()


async def get_all_industries() -> List:
    """Get all industries."""
    return await find_all_industries()


async def create_industry(name: str):
    """Create a new industry with normalized name, or return existing."""
    normalized_name = normalize_industry_name(name)

    existing = await find_industry_by_name(normalized_name)
    if existing:
        return existing

    return await create_industry_repo(normalized_name)

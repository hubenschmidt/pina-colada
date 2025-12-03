"""Service layer for technology business logic."""

from typing import Optional, List

from fastapi import HTTPException

from repositories.technology_repository import (
    find_all_technologies,
    find_technology_by_id,
    create_technology as create_technology_repo,
)


async def get_all_technologies(category: Optional[str] = None) -> List:
    """Get all technologies, optionally filtered by category."""
    return await find_all_technologies(category=category)


async def get_technology(tech_id: int):
    """Get a single technology by ID."""
    tech = await find_technology_by_id(tech_id)
    if not tech:
        raise HTTPException(status_code=404, detail="Technology not found")
    return tech


async def create_technology(name: str, category: str, vendor: Optional[str] = None):
    """Create a new technology."""
    return await create_technology_repo(name=name, category=category, vendor=vendor)

"""Service layer for funding stage business logic."""

from typing import List

from repositories.funding_stage_repository import find_all_funding_stages


async def get_all_funding_stages() -> List:
    """Get all funding stages."""
    return await find_all_funding_stages()

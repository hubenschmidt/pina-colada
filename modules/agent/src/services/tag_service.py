"""Service layer for tag business logic."""

from typing import List

from repositories.tag_repository import find_all_tags


async def get_all_tags() -> List:
    """Get all tags ordered by name."""
    return await find_all_tags()

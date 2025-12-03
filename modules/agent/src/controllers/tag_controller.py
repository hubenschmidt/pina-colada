"""Controller layer for tag routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from services.tag_service import get_all_tags as get_all_tags_service


@handle_http_exceptions
async def get_all_tags() -> List[dict]:
    """Get all tags."""
    tags = await get_all_tags_service()
    return [{"id": tag.id, "name": tag.name} for tag in tags]

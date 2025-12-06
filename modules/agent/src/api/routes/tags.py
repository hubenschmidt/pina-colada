"""Routes for tags API endpoints."""

from fastapi import APIRouter, Request

from controllers.tag_controller import get_all_tags


router = APIRouter(prefix="/tags", tags=["tags"])


@router.get("")
async def list_tags_route(request: Request):
    """List all tags."""
    return await get_all_tags()

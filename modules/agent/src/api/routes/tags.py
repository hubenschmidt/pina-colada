"""Routes for tags API endpoints."""

from fastapi import APIRouter, Request

from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.tag_controller import get_all_tags


router = APIRouter(prefix="/tags", tags=["tags"])


@router.get("")
@require_auth
@log_errors
async def list_tags_route(request: Request):
    """List all tags."""
    return await get_all_tags()

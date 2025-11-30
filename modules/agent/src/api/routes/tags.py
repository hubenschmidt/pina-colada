"""Routes for tags API endpoints."""

from fastapi import APIRouter, Request
from sqlalchemy import select

from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.db import async_get_session
from models.Tag import Tag


router = APIRouter(prefix="/tags", tags=["tags"])


@router.get("")
@require_auth
@log_errors
async def list_tags_route(request: Request):
    """List all tags."""
    async with async_get_session() as session:
        stmt = select(Tag).order_by(Tag.name)
        result = await session.execute(stmt)
        tags = result.scalars().all()
        return [{"id": tag.id, "name": tag.name} for tag in tags]

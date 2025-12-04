"""Routes for comments API endpoints."""

from fastapi import APIRouter, Request

from controllers.comment_controller import (
    create_comment,
    delete_comment,
    get_comment,
    get_comments,
    update_comment,
)
from schemas.comment import CommentCreate, CommentUpdate
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/comments", tags=["comments"])


@router.get("")
@require_auth
@log_errors
async def get_comments_route(request: Request, commentable_type: str, commentable_id: int):
    """Get all comments for a specific entity."""
    return await get_comments(request, commentable_type, commentable_id)


@router.get("/{comment_id}")
@require_auth
@log_errors
async def get_comment_route(request: Request, comment_id: int):
    """Get a single comment by ID."""
    return await get_comment(request, comment_id)


@router.post("")
@require_auth
@log_errors
async def create_comment_route(request: Request, data: CommentCreate):
    """Create a new comment."""
    return await create_comment(request, data)


@router.put("/{comment_id}")
@require_auth
@log_errors
async def update_comment_route(request: Request, comment_id: int, data: CommentUpdate):
    """Update a comment."""
    return await update_comment(request, comment_id, data)


@router.delete("/{comment_id}")
@require_auth
@log_errors
async def delete_comment_route(request: Request, comment_id: int):
    """Delete a comment."""
    return await delete_comment(request, comment_id)

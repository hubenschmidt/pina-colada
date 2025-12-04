"""Controller layer for comment routing to services."""

from typing import List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.comment import comment_to_response
from services.comment_service import (
    CommentCreate,
    CommentUpdate,
    get_comments_by_entity as get_comments_service,
    get_comment as get_comment_service,
    create_comment as create_comment_service,
    update_comment as update_comment_service,
    delete_comment as delete_comment_service,
)

# Re-export for routes
__all__ = ["CommentCreate", "CommentUpdate"]


@handle_http_exceptions
async def get_comments(request: Request, commentable_type: str, commentable_id: int) -> List[dict]:
    """Get all comments for a specific entity."""
    tenant_id = request.state.tenant_id
    comments = await get_comments_service(tenant_id, commentable_type, commentable_id)
    return [comment_to_response(c) for c in comments]


@handle_http_exceptions
async def get_comment(request: Request, comment_id: int) -> dict:
    """Get a single comment by ID."""
    tenant_id = request.state.tenant_id
    comment = await get_comment_service(comment_id, tenant_id)
    return comment_to_response(comment)


@handle_http_exceptions
async def create_comment(request: Request, data: CommentCreate) -> dict:
    """Create a new comment."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    comment = await create_comment_service(
        tenant_id, data.commentable_type, data.commentable_id,
        data.content, user_id, data.parent_comment_id
    )
    return comment_to_response(comment)


@handle_http_exceptions
async def update_comment(request: Request, comment_id: int, data: CommentUpdate) -> dict:
    """Update a comment."""
    tenant_id = request.state.tenant_id
    comment = await update_comment_service(comment_id, tenant_id, data.content)
    return comment_to_response(comment)


@handle_http_exceptions
async def delete_comment(request: Request, comment_id: int) -> dict:
    """Delete a comment."""
    tenant_id = request.state.tenant_id
    await delete_comment_service(comment_id, tenant_id)
    return {"status": "deleted"}

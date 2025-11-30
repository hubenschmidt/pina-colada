"""Routes for comments API endpoints."""

from fastapi import APIRouter, Request, HTTPException
from pydantic import BaseModel

from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.comment_repository import (
    find_comments_by_entity,
    find_comment_by_id,
    create_comment,
    update_comment,
    delete_comment,
)
from services.notification_service import create_comment_notifications


# Normalize commentable_type to PascalCase for consistency with database
COMMENTABLE_TYPE_MAP = {
    "account": "Account",
    "lead": "Lead",
    "deal": "Deal",
    "task": "Task",
}


def _normalize_commentable_type(commentable_type: str) -> str:
    """Normalize commentable_type to PascalCase."""
    return COMMENTABLE_TYPE_MAP.get(commentable_type.lower(), commentable_type)


class CommentCreate(BaseModel):
    commentable_type: str
    commentable_id: int
    content: str
    parent_comment_id: int | None = None


class CommentUpdate(BaseModel):
    content: str


def _comment_to_dict(comment) -> dict:
    """Convert Comment model to dictionary."""
    creator = comment.creator
    created_by_name = None
    created_by_email = None
    if creator:
        first = creator.first_name or ""
        last = creator.last_name or ""
        created_by_name = f"{first} {last}".strip() or None
        created_by_email = creator.email

    return {
        "id": comment.id,
        "tenant_id": comment.tenant_id,
        "commentable_type": comment.commentable_type,
        "commentable_id": comment.commentable_id,
        "content": comment.content,
        "created_by": comment.created_by,
        "created_by_name": created_by_name,
        "created_by_email": created_by_email,
        "parent_comment_id": comment.parent_comment_id,
        "created_at": comment.created_at.isoformat() if comment.created_at else None,
        "updated_at": comment.updated_at.isoformat() if comment.updated_at else None,
    }


router = APIRouter(prefix="/comments", tags=["comments"])


@router.get("")
@require_auth
@log_errors
async def get_comments_route(
    request: Request,
    commentable_type: str,
    commentable_id: int,
):
    """Get all comments for a specific entity."""
    tenant_id = request.state.tenant_id
    normalized_type = _normalize_commentable_type(commentable_type)
    comments = await find_comments_by_entity(tenant_id, normalized_type, commentable_id)
    return [_comment_to_dict(c) for c in comments]


@router.get("/{comment_id}")
@require_auth
@log_errors
async def get_comment_route(request: Request, comment_id: int):
    """Get a single comment by ID."""
    tenant_id = request.state.tenant_id
    comment = await find_comment_by_id(comment_id, tenant_id)
    if not comment:
        raise HTTPException(status_code=404, detail="Comment not found")
    return _comment_to_dict(comment)


@router.post("")
@require_auth
@log_errors
async def create_comment_route(request: Request, data: CommentCreate):
    """Create a new comment."""
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)

    if not data.content.strip():
        raise HTTPException(status_code=400, detail="Comment content cannot be empty")

    comment_data = {
        "tenant_id": tenant_id,
        "commentable_type": _normalize_commentable_type(data.commentable_type),
        "commentable_id": data.commentable_id,
        "content": data.content,
        "created_by": user_id,
        "parent_comment_id": data.parent_comment_id,
    }

    comment = await create_comment(comment_data)

    # Create notifications for relevant users (async, fire-and-forget)
    try:
        await create_comment_notifications(comment)
    except Exception as e:
        # Log but don't fail the comment creation
        import logging
        logging.getLogger(__name__).error(f"Failed to create notifications: {e}")

    return _comment_to_dict(comment)


@router.put("/{comment_id}")
@require_auth
@log_errors
async def update_comment_route(request: Request, comment_id: int, data: CommentUpdate):
    """Update a comment."""
    tenant_id = request.state.tenant_id

    if not data.content.strip():
        raise HTTPException(status_code=400, detail="Comment content cannot be empty")

    comment = await update_comment(comment_id, tenant_id, {"content": data.content})
    if not comment:
        raise HTTPException(status_code=404, detail="Comment not found")
    return _comment_to_dict(comment)


@router.delete("/{comment_id}")
@require_auth
@log_errors
async def delete_comment_route(request: Request, comment_id: int):
    """Delete a comment."""
    tenant_id = request.state.tenant_id

    deleted = await delete_comment(comment_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Comment not found")
    return {"status": "deleted"}

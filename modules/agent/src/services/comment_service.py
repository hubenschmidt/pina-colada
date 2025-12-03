"""Service layer for comment business logic."""

import logging
from typing import Optional, Dict, Any, List

from fastapi import HTTPException

from repositories.comment_repository import (
    find_comments_by_entity,
    find_comment_by_id,
    create_comment as create_comment_repo,
    update_comment as update_comment_repo,
    delete_comment as delete_comment_repo,
)
from services.notification_service import create_comment_notifications

logger = logging.getLogger(__name__)

COMMENTABLE_TYPE_MAP = {
    "account": "Account",
    "lead": "Lead",
    "deal": "Deal",
    "task": "Task",
}


def _normalize_commentable_type(commentable_type: str) -> str:
    """Normalize commentable_type to PascalCase."""
    return COMMENTABLE_TYPE_MAP.get(commentable_type.lower(), commentable_type)


async def get_comments_by_entity(tenant_id: int, commentable_type: str, commentable_id: int) -> List:
    """Get all comments for a specific entity."""
    normalized_type = _normalize_commentable_type(commentable_type)
    return await find_comments_by_entity(tenant_id, normalized_type, commentable_id)


async def get_comment(comment_id: int, tenant_id: int):
    """Get a single comment by ID."""
    comment = await find_comment_by_id(comment_id, tenant_id)
    if not comment:
        raise HTTPException(status_code=404, detail="Comment not found")
    return comment


async def create_comment(
    tenant_id: int,
    commentable_type: str,
    commentable_id: int,
    content: str,
    user_id: Optional[int],
    parent_comment_id: Optional[int] = None,
):
    """Create a new comment and trigger notifications."""
    if not content.strip():
        raise HTTPException(status_code=400, detail="Comment content cannot be empty")

    comment_data = {
        "tenant_id": tenant_id,
        "commentable_type": _normalize_commentable_type(commentable_type),
        "commentable_id": commentable_id,
        "content": content,
        "created_by": user_id,
        "parent_comment_id": parent_comment_id,
    }

    comment = await create_comment_repo(comment_data)

    try:
        await create_comment_notifications(comment)
    except Exception as e:
        logger.error(f"Failed to create notifications: {e}")

    return comment


async def update_comment(comment_id: int, tenant_id: int, content: str):
    """Update a comment."""
    if not content.strip():
        raise HTTPException(status_code=400, detail="Comment content cannot be empty")

    comment = await update_comment_repo(comment_id, tenant_id, {"content": content})
    if not comment:
        raise HTTPException(status_code=404, detail="Comment not found")
    return comment


async def delete_comment(comment_id: int, tenant_id: int):
    """Delete a comment."""
    deleted = await delete_comment_repo(comment_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Comment not found")
    return True

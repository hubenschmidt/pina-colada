"""Service layer for comment notifications."""

import logging
from typing import TYPE_CHECKING, Optional, List

if TYPE_CHECKING:
    from models.Comment import Comment

from repositories.notification_repository import (
    create_notification,
    get_thread_participants,
    get_unread_count as repo_get_unread_count,
    get_notifications as repo_get_notifications,
    mark_as_read as repo_mark_as_read,
    mark_entity_as_read as repo_mark_entity_as_read,
    get_entity_project_id_sync,
    get_lead_type_sync,
    MarkReadRequest,
    MarkEntityReadRequest,
)
from repositories.comment_repository import find_comment_by_id

# Re-export Pydantic models for controllers
__all__ = ["MarkReadRequest", "MarkEntityReadRequest"]

logger = logging.getLogger(__name__)

ENTITY_URL_MAP = {
    "Lead": "/leads",
    "Deal": "/deals",
    "Task": "/tasks",
    "Individual": "/accounts/individuals",
    "Organization": "/accounts/organizations",
}


def get_entity_url(entity_type: str, entity_id: int, comment_id: Optional[int] = None) -> str:
    """Get URL for an entity, optionally with comment anchor."""
    if entity_type == "Lead":
        lead_type = get_lead_type_sync(entity_id)
        url = f"/leads/{lead_type}/{entity_id}"
    else:
        base_path = ENTITY_URL_MAP.get(entity_type, f"/{entity_type.lower()}s")
        url = f"{base_path}/{entity_id}"

    if not comment_id:
        return url
    return f"{url}#comment-{comment_id}"


def get_entity_project_id(entity_type: str, entity_id: int) -> Optional[int]:
    """Public accessor for entity project lookup."""
    return get_entity_project_id_sync(entity_type, entity_id)


async def get_unread_count(user_id: int, tenant_id: int) -> int:
    """Get unread notification count for a user."""
    return await repo_get_unread_count(user_id, tenant_id)


async def get_notifications(user_id: int, tenant_id: int, limit: int) -> List:
    """Get notifications for a user."""
    return await repo_get_notifications(user_id, tenant_id, limit)


async def mark_as_read(notification_ids: List[int], user_id: int, tenant_id: int) -> int:
    """Mark specific notifications as read."""
    return await repo_mark_as_read(notification_ids, user_id, tenant_id)


async def mark_entity_as_read(user_id: int, tenant_id: int, entity_type: str, entity_id: int) -> int:
    """Mark all notifications for an entity as read."""
    return await repo_mark_entity_as_read(user_id, tenant_id, entity_type, entity_id)


async def create_comment_notifications(comment: "Comment") -> None:
    """
    Create notifications when a comment is created.

    - direct_reply: Notify the author of the parent comment (if this is a reply)
    - thread_activity: Notify all users who have commented on the same entity
    """
    if not comment.created_by:
        # Anonymous comment, no notifications
        return

    notified_users: set[int] = set()

    # 1. Direct reply notification
    if comment.parent_comment_id:
        parent = await find_comment_by_id(comment.parent_comment_id, comment.tenant_id)
        if parent and parent.created_by and parent.created_by != comment.created_by:
            await create_notification(
                tenant_id=comment.tenant_id,
                user_id=parent.created_by,
                comment_id=comment.id,
                notification_type="direct_reply",
            )
            notified_users.add(parent.created_by)
            logger.debug(f"Created direct_reply notification for user {parent.created_by}")

    # 2. Thread activity notifications
    participants = await get_thread_participants(
        tenant_id=comment.tenant_id,
        entity_type=comment.commentable_type,
        entity_id=comment.commentable_id,
        exclude_user_id=comment.created_by,
    )

    unnotified_participants = [uid for uid in participants if uid not in notified_users]
    for user_id in unnotified_participants:
        await create_notification(
            tenant_id=comment.tenant_id,
            user_id=user_id,
            comment_id=comment.id,
            notification_type="thread_activity",
        )
        logger.debug(f"Created thread_activity notification for user {user_id}")

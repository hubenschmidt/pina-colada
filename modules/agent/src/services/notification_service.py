"""Service layer for comment notifications."""

import logging
from typing import Optional

from models.Comment import Comment
from repositories.notification_repository import (
    create_notification,
    get_thread_participants,
)
from repositories.comment_repository import find_comment_by_id

logger = logging.getLogger(__name__)


async def create_comment_notifications(comment: Comment) -> None:
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
